import { cleanup, fireEvent, screen, waitFor } from '@testing-library/react';
import { afterEach, beforeAll, describe, expect, it, vi } from 'vitest';
import { json, me, renderApp, setupTestApp, status, stubApi, version } from '../../../test-support';

beforeAll(setupTestApp);

afterEach(() => {
  cleanup();
  vi.unstubAllGlobals();
});

const dashboard = {
  today: '2026-10-19',
  streak: 0,
  studied_today: false,
  today_stats: { reviews: 0, correct: 0, new_items: 0, active_seconds: 0 },
  forecast: [],
  known_words: 0,
  most_missed: [],
};
const heatmap = { from: '2026-07-28', to: '2026-10-19', days: [] };
const suggestions = {
  based_on_history: true,
  suggestions: [
    { preset: 'light', daily_minutes: 5, days_target: 2 },
    { preset: 'steady', daily_minutes: 10, days_target: 3 },
    { preset: 'intense', daily_minutes: 20, days_target: 4 },
  ],
};
const goal = {
  id: 'g1',
  period_start: '2026-10-19',
  period_end: '2026-10-25',
  preset: 'steady',
  daily_minutes: 10,
  days_target: 5,
  weekdays: [1, 2, 3, 4, 5, 6, 7],
};
const progress = {
  days_studied: 3,
  days_target: 5,
  minutes_studied: 45,
  minutes_goal: 50,
  new_words: 6,
  days_left: 3,
  met: false,
};
const active = { goal, progress, ended: false };
const ended = {
  goal,
  progress: { ...progress, days_studied: 5, days_left: 0, met: true, minutes_studied: 62 },
  ended: true,
};

type Setup = { current?: unknown; saved?: () => Response; suggested?: unknown };

function api({ current = { ended: false }, saved = () => json(active, 201), suggested = suggestions }: Setup = {}) {
  return stubApi({
    'GET /version': () => json(version),
    'GET /me': () => json(me),
    'GET /progress/dashboard': () => json(dashboard),
    'GET /progress/heatmap': () => json(heatmap),
    'GET /goals/current': () => (current === 'error' ? status(500) : json(current)),
    'GET /goals/suggestions': () => json(suggested),
    'POST /goals': () => saved(),
  });
}

const posted = (calls: ReturnType<typeof api>) =>
  calls.filter((call) => call.method === 'POST' && call.path === '/goals');

describe('cartão da meta no Início', () => {
  it('sem meta, convida a definir e leva à tela de metas', async () => {
    api();
    renderApp('/');
    expect(await screen.findByText('Defina sua meta da semana')).toBeTruthy();
    fireEvent.click(screen.getByRole('link', { name: 'Definir meta' }));
    expect(await screen.findByRole('heading', { name: 'Metas' })).toBeTruthy();
  });

  it('com meta, mostra dias, minutos, o que falta e a barra', async () => {
    api({ current: active });
    renderApp('/');
    expect(await screen.findByText('3 de 5 dias estudados')).toBeTruthy();
    expect(screen.getByText('45 de 50 minutos')).toBeTruthy();
    expect(screen.getByText('Faltam 3 dias')).toBeTruthy();
    expect(screen.getByRole('progressbar', { name: 'Dias estudados no ciclo' }).getAttribute('aria-valuenow')).toBe(
      '60',
    );
    expect(screen.getByRole('link', { name: 'Ver meta' })).toBeTruthy();
  });

  it('no último dia do ciclo, avisa', async () => {
    api({ current: { ...active, progress: { ...progress, days_left: 0 } } });
    renderApp('/');
    expect(await screen.findByText('Hoje é o último dia do ciclo')).toBeTruthy();
  });

  it('acabado o ciclo, leva à retrospectiva e à próxima meta', async () => {
    api({ current: ended });
    renderApp('/');
    expect(await screen.findByText('Seu ciclo acabou')).toBeTruthy();
    expect(screen.getByText('Meta cumprida!')).toBeTruthy();
    expect(screen.getByRole('link', { name: 'Ver como foi e escolher a próxima' })).toBeTruthy();
  });

  it('se o servidor falhar, não mostra o cartão e o Início continua', async () => {
    api({ current: 'error' });
    renderApp('/');
    expect(await screen.findByRole('link', { name: /Estudar agora/ })).toBeTruthy();
    await waitFor(() => expect(screen.queryByText('Defina sua meta da semana')).toBeNull());
  });
});

describe('tela de metas', () => {
  it('escolher uma sugestão manda os valores dela e volta ao Início', async () => {
    const calls = api();
    renderApp('/goals');
    expect(await screen.findByText('Sugestões baseadas no seu ritmo das últimas 4 semanas.')).toBeTruthy();
    expect(screen.getByRole('heading', { name: 'Defina sua meta da semana' })).toBeTruthy();
    fireEvent.click(screen.getByRole('radio', { name: /Intensa/ }));
    fireEvent.click(screen.getByRole('button', { name: 'Começar' }));

    await waitFor(() => expect(posted(calls)).toHaveLength(1));
    expect(posted(calls)[0]?.body).toEqual({ preset: 'intense', daily_minutes: 20, days_target: 4 });
    expect(posted(calls)[0]?.headers.get('X-Requested-With')).toBe('XMLHttpRequest');
    expect(await screen.findByRole('link', { name: /Estudar agora/ })).toBeTruthy();
  });

  it('a sugestão "Constante" já vem escolhida', async () => {
    const calls = api();
    renderApp('/goals');
    fireEvent.click(await screen.findByRole('button', { name: 'Começar' }));
    await waitFor(() => expect(posted(calls)).toHaveLength(1));
    expect(posted(calls)[0]?.body).toMatchObject({ preset: 'steady', daily_minutes: 10, days_target: 3 });
  });

  it('sem histórico, avisa que as sugestões são para começar', async () => {
    api({ suggested: { ...suggestions, based_on_history: false } });
    renderApp('/goals');
    expect(await screen.findByText(/Sugestões para começar/)).toBeTruthy();
  });

  it('personalizar manda o que foi digitado', async () => {
    const calls = api();
    renderApp('/goals');
    fireEvent.click(await screen.findByRole('radio', { name: 'Personalizar' }));
    fireEvent.change(screen.getByLabelText('Minutos por dia'), { target: { value: '25' } });
    fireEvent.change(screen.getByLabelText('Dias por semana'), { target: { value: '6' } });
    fireEvent.click(screen.getByRole('button', { name: 'Começar' }));
    await waitFor(() => expect(posted(calls)).toHaveLength(1));
    expect(posted(calls)[0]?.body).toEqual({ preset: 'custom', daily_minutes: 25, days_target: 6 });
  });

  it('personalizar com valor fora do limite mostra o erro e não deixa começar', async () => {
    const calls = api();
    renderApp('/goals');
    fireEvent.click(await screen.findByRole('radio', { name: 'Personalizar' }));
    fireEvent.change(screen.getByLabelText('Minutos por dia'), { target: { value: '500' } });
    fireEvent.change(screen.getByLabelText('Dias por semana'), { target: { value: '9' } });
    expect(screen.getByText('Use de 1 a 240 minutos')).toBeTruthy();
    expect(screen.getByText('Use de 1 a 7 dias')).toBeTruthy();
    const start = screen.getByRole('button', { name: 'Começar' });
    expect((start as HTMLButtonElement).disabled).toBe(true);
    fireEvent.click(start);
    expect(posted(calls)).toHaveLength(0);
  });
});

describe('meta em andamento, retrospectiva e falhas na tela de metas', () => {
  it('com a meta em andamento mostra o progresso e "Mudar a meta", sem "Manter a mesma"', async () => {
    api({ current: active });
    renderApp('/goals');
    expect(await screen.findByRole('heading', { name: 'Mudar a meta' })).toBeTruthy();
    expect(screen.getByText('3 de 5 dias estudados')).toBeTruthy();
    expect(screen.queryByRole('button', { name: 'Manter a mesma' })).toBeNull();
  });

  it('acabado o ciclo, mostra a retrospectiva e deixa manter a mesma meta', async () => {
    const calls = api({ current: ended });
    renderApp('/goals');
    expect(await screen.findByText('Meta cumprida! Você estudou 5 de 5 dias.')).toBeTruthy();
    expect(screen.getByRole('heading', { name: 'Qual a meta da próxima semana?' })).toBeTruthy();
    expect(screen.getByText('62')).toBeTruthy();
    expect(screen.getByText('6')).toBeTruthy();
    fireEvent.click(screen.getByRole('button', { name: 'Manter a mesma' }));
    await waitFor(() => expect(posted(calls)).toHaveLength(1));
    expect(posted(calls)[0]?.body).toEqual({
      preset: 'steady',
      daily_minutes: 10,
      days_target: 5,
      weekdays: [1, 2, 3, 4, 5, 6, 7],
    });
  });

  it('ciclo não cumprido tem retrospectiva com incentivo', async () => {
    api({ current: { ...ended, progress: { ...ended.progress, days_studied: 2, met: false } } });
    renderApp('/goals');
    expect(await screen.findByText('Você estudou 2 de 5 dias. Dá para tentar de novo!')).toBeTruthy();
  });

  it('se salvar falhar, avisa e fica na tela', async () => {
    api({ saved: () => status(500) });
    renderApp('/goals');
    fireEvent.click(await screen.findByRole('button', { name: 'Começar' }));
    expect(await screen.findByText('Não consegui salvar sua meta agora. Tente de novo.')).toBeTruthy();
    expect(screen.getByRole('heading', { name: 'Metas' })).toBeTruthy();
  });

  it('se não carregar, avisa', async () => {
    api({ current: 'error' });
    renderApp('/goals');
    expect(await screen.findByText('Não consegui carregar suas metas agora.')).toBeTruthy();
  });
});
