import { cleanup, fireEvent, screen, waitFor, within } from '@testing-library/react';
import { afterEach, beforeAll, describe, expect, it } from 'vitest';
import { json, me, renderApp, setupTestApp, status, stubApi, version } from '../../../test-support';

beforeAll(setupTestApp);
afterEach(cleanup);

const week = {
  period: 'week',
  from: '2026-11-23',
  to: '2026-11-29',
  entries: [
    { position: 1, nickname: 'bia', points: 420, me: false },
    { position: 2, nickname: 'caio', points: 300, me: false },
    { position: 3, nickname: 'duda', points: 150, me: false },
    { position: 4, nickname: 'eva', points: 90, me: false },
  ],
  me: { position: 6, points: 40, me: true },
};
const all = {
  ...week,
  period: 'all',
  entries: [{ position: 1, nickname: 'ana', points: 9000, me: true }],
  me: undefined,
};
const score = { week_points: 120, total_points: 260, level: 3, level_floor: 250, next_level: 450, week_days_met: 2 };
const dashboard = {
  today: '2026-11-25',
  streak: 0,
  studied_today: false,
  today_stats: { reviews: 0, correct: 0, new_items: 0, active_seconds: 0 },
  forecast: [],
  known_words: 0,
  most_missed: [],
};

type Setup = { saved?: () => Response; profile?: unknown };

function api({
  saved = () => json({ nickname: 'ana_1', visible: true }),
  profile = { nickname: '', visible: false },
}: Setup = {}) {
  return stubApi({
    'GET /version': () => json(version),
    'GET /me': () => json(me),
    'GET /leaderboard': (req: Request) => json(new URL(req.url).searchParams.get('period') === 'all' ? all : week),
    'GET /leaderboard/me': () => json(score),
    'GET /leaderboard/profile': () => json(profile),
    'PUT /leaderboard/profile': () => saved(),
    'GET /progress/dashboard': () => json(dashboard),
    'GET /progress/heatmap': () => json({ from: '2026-09-01', to: '2026-11-25', days: [] }),
    'GET /goals/current': () => json({ ended: false }),
  });
}

describe('tela de ranking', () => {
  it('mostra o pódio, a lista e a minha linha (só minha) fora do top', async () => {
    api();
    renderApp('/ranking');
    const podium = await screen.findByRole('list', { name: 'Pódio' });
    expect(within(podium).getByText('bia')).toBeTruthy();
    expect(within(podium).getByText('420 pontos')).toBeTruthy();
    expect(within(screen.getByRole('list', { name: 'Demais posições' })).getByText('eva')).toBeTruthy();
    const mine = screen.getByRole('region', { name: 'Sua posição' });
    expect(within(mine).getByText('6º')).toBeTruthy();
    expect(within(mine).getByText('Você')).toBeTruthy();
    expect(screen.getByText(/Só você vê esta linha/)).toBeTruthy();
  });

  it('a aba Geral troca o período e marca quem sou no pódio', async () => {
    api();
    renderApp('/ranking');
    fireEvent.click(await screen.findByRole('button', { name: 'Geral' }));
    expect(await screen.findByText('ana (você)')).toBeTruthy();
  });

  it('participar manda o apelido e a participação', async () => {
    const calls = api();
    renderApp('/ranking');
    fireEvent.change(await screen.findByLabelText('Apelido'), { target: { value: 'ana_1' } });
    fireEvent.click(screen.getByRole('button', { name: 'Salvar' }));
    expect(await screen.findByText('Salvo!')).toBeTruthy();
    const put = calls.find((call) => call.method === 'PUT' && call.path === '/leaderboard/profile');
    expect(put?.body).toEqual({ nickname: 'ana_1', visible: true });
    expect(put?.headers.get('X-Requested-With')).toBe('XMLHttpRequest');
  });

  it('apelido em uso (409) avisa; apelido fora do formato não deixa salvar', async () => {
    api({ saved: () => status(409) });
    renderApp('/ranking');
    const field = await screen.findByLabelText('Apelido');
    fireEvent.change(field, { target: { value: 'ab' } });
    expect((screen.getByRole('button', { name: 'Salvar' }) as HTMLButtonElement).disabled).toBe(true);
    fireEvent.change(field, { target: { value: 'bia' } });
    fireEvent.click(screen.getByRole('button', { name: 'Salvar' }));
    expect(await screen.findByText('Esse apelido já está em uso. Escolha outro.')).toBeTruthy();
  });
});

describe('cartão de pontos no Início', () => {
  it('mostra os pontos da semana, o nível com a barra e leva ao ranking', async () => {
    api();
    renderApp('/');
    expect(await screen.findByText('120 pontos nesta semana')).toBeTruthy();
    expect(screen.getByText('Faltam 190 pontos para o nível 4')).toBeTruthy();
    expect(screen.getByText('Meta cumprida em 2 de 7 dias')).toBeTruthy();
    expect(screen.getByRole('progressbar', { name: 'Progresso do nível' }).getAttribute('aria-valuenow')).toBe('5');
    fireEvent.click(screen.getByRole('link', { name: 'Ver o ranking' }));
    await waitFor(() => expect(screen.getByRole('heading', { name: 'Ranking' })).toBeTruthy());
  });
});
