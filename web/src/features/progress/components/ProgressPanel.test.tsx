import { cleanup, screen, waitFor } from '@testing-library/react';
import { afterEach, beforeAll, describe, expect, it, vi } from 'vitest';
import { json, me, renderApp, setupTestApp, status, stubApi, version } from '../../../test-support';

beforeAll(setupTestApp);

afterEach(() => {
  cleanup();
  vi.unstubAllGlobals();
});

const dashboard = {
  today: '2026-10-19',
  streak: 5,
  studied_today: true,
  accuracy_7d: 0.784,
  today_stats: { reviews: 12, correct: 9, new_items: 3, active_seconds: 140 },
  forecast: [
    { date: '2026-10-19', due: 9 },
    { date: '2026-10-20', due: 0 },
    { date: '2026-10-21', due: 3 },
    { date: '2026-10-22', due: 0 },
    { date: '2026-10-23', due: 0 },
    { date: '2026-10-24', due: 0 },
    { date: '2026-10-25', due: 0 },
  ],
  known_words: 42,
  most_missed: [
    { item_id: 'i1', term: 'house', gloss: 'casa', misses: 3 },
    { item_id: 'i2', term: 'mother', gloss: 'mãe', misses: 1 },
  ],
};
const heatmap = {
  from: '2026-07-28',
  to: '2026-10-19',
  days: [
    { date: '2026-10-18', reviews: 10, correct: 8 },
    { date: '2026-10-19', reviews: 12, correct: 9 },
  ],
};
const nothing = {
  ...dashboard,
  streak: 0,
  studied_today: false,
  accuracy_7d: undefined,
  today_stats: { reviews: 0, correct: 0, new_items: 0, active_seconds: 0 },
  known_words: 0,
  most_missed: [],
};

function home(board: unknown, map: unknown, reply = json) {
  return stubApi({
    'GET /version': () => json(version),
    'GET /me': () => json(me),
    'GET /progress/dashboard': () => reply(board),
    'GET /progress/heatmap': () => reply(map),
  });
}

describe('painel de progresso no Início', () => {
  it('mostra a sequência, os acertos, as palavras, o heatmap, a previsão e as mais erradas', async () => {
    home(dashboard, heatmap);
    renderApp('/');
    expect(await screen.findByText('5 dias seguidos')).toBeTruthy();
    expect(screen.getByText('Hoje já conta!')).toBeTruthy();
    expect(screen.getByText('12 respostas · 3 palavras novas')).toBeTruthy();
    expect(screen.getByText('78%')).toBeTruthy();
    expect(screen.getByText('42')).toBeTruthy();
    expect(screen.getByRole('img', { name: 'Você estudou em 2 dias nas últimas 12 semanas' })).toBeTruthy();
    expect(screen.getByText('12 revisões nos próximos 7 dias')).toBeTruthy();
    expect(screen.getByText('house')).toBeTruthy();
    expect(screen.getByText('3 erros')).toBeTruthy();
    expect(screen.getByText('1 erro')).toBeTruthy();
  });

  it('o heatmap tem 12 semanas de células, com tons conforme as respostas', async () => {
    home(dashboard, heatmap);
    const { container } = renderApp('/');
    await screen.findByText('5 dias seguidos');
    expect(container.querySelectorAll('[data-level]')).toHaveLength(84);
    expect(container.querySelectorAll('[data-level="2"]')).toHaveLength(2); // 10 e 12 respostas: tom 2
    expect(container.querySelectorAll('[data-level="0"]')).toHaveLength(82);
  });

  it('os atalhos de estudar e de palavras continuam no Início', async () => {
    home(dashboard, heatmap);
    renderApp('/');
    expect(await screen.findByRole('link', { name: /Estudar agora/ })).toBeTruthy();
    expect(screen.getByRole('link', { name: /Escolha uma unidade/ })).toBeTruthy();
  });

  it('hoje ainda não estudou: a sequência pede para manter', async () => {
    home({ ...dashboard, studied_today: false }, heatmap);
    renderApp('/');
    expect(await screen.findByText('Estude hoje para manter a sequência.')).toBeTruthy();
    expect(screen.queryByText('Hoje já conta!')).toBeNull();
  });

  it('sem respostas nos 7 dias não inventa porcentagem', async () => {
    home({ ...dashboard, accuracy_7d: undefined }, heatmap);
    renderApp('/');
    expect(await screen.findByText('Sem respostas ainda')).toBeTruthy();
    expect(screen.queryByText(/%$/)).toBeNull();
  });

  it('quem nunca estudou vê só o convite', async () => {
    home(nothing, { ...heatmap, days: [] });
    renderApp('/');
    expect(await screen.findByText('Faça uma sessão de estudo para ver seu progresso aqui.')).toBeTruthy();
    expect(screen.queryByText('Sequência')).toBeNull();
  });

  it('sem nada para revisar nem errar, avisa', async () => {
    home({ ...dashboard, forecast: dashboard.forecast.map((day) => ({ ...day, due: 0 })), most_missed: [] }, heatmap);
    renderApp('/');
    expect(await screen.findByText('Nada para revisar nos próximos 7 dias.')).toBeTruthy();
    expect(screen.getByText('Nenhum erro no último mês.')).toBeTruthy();
  });

  it('se o servidor falhar, avisa sem derrubar o Início', async () => {
    home(dashboard, heatmap, () => status(500));
    renderApp('/');
    expect(await screen.findByText('Não consegui carregar seu progresso agora.')).toBeTruthy();
    await waitFor(() => expect(screen.getByRole('link', { name: /Estudar agora/ })).toBeTruthy());
  });
});
