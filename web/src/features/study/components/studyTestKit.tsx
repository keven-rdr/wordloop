import { fireEvent, screen } from '@testing-library/react';
import { type Call, json, me, renderApp, stubApi, version } from '../../../test-support';
import { setStoreForTests } from '../offline/store';

// Cartões e atalhos compartilhados pelos testes da sessão de estudo.
export const house = {
  id: 'c1',
  item_id: 'i1',
  direction: 'receptive',
  state: 'new',
  exercise: 'choose',
  term: 'house',
  gloss: 'casa',
  mastery: 0,
  options: [
    { item_id: 'i2', gloss: 'mãe' },
    { item_id: 'i1', gloss: 'casa' },
  ],
};
export const mother = {
  ...house,
  id: 'c2',
  item_id: 'i2',
  exercise: 'type',
  term: 'mother',
  gloss: 'mãe',
  options: undefined,
};
export const pack = (cards: unknown[]) => ({
  session_id: 'sess-1',
  cards,
  new_limit: 8,
  review_limit: 20,
  expires_at: '2999-01-01T00:00:00Z', // longe no futuro: o pacote nunca vence durante os testes
});
export const accepted = (req: Request) =>
  req
    .clone()
    .json()
    .then((body: { reviews: { id: string }[] }) =>
      json({
        results: body.reviews.map((review) => ({ id: review.id, status: 'applied', correct: true })),
      }),
    );

// A meta da semana em andamento: o tempo da sessão vem dela (15 minutos por dia).
export const weeklyGoal = {
  goal: {
    id: 'g1',
    period_start: '2026-11-16',
    period_end: '2026-11-22',
    preset: 'steady',
    daily_minutes: 15,
    days_target: 7,
    weekdays: [1, 2, 3, 4, 5, 6, 7],
  },
  progress: {
    days_studied: 1,
    days_target: 7,
    minutes_studied: 5,
    minutes_goal: 105,
    new_words: 3,
    days_left: 4,
    met: false,
  },
  ended: false,
};

export function api(cards: unknown[], reviews: (req: Request) => Response | Promise<Response> = accepted) {
  setStoreForTests(); // cada teste começa com a fila do aparelho vazia
  return stubApi({
    'GET /version': () => json(version),
    'GET /me': () => json(me),
    'GET /goals/current': () => json(weeklyGoal),
    'POST /study/plan': () => json(pack(cards)),
    'POST /study/reviews': reviews,
  });
}

// Respostas do primeiro lote enviado ao servidor.
export function sentReviews(calls: Call[]): Record<string, unknown>[] {
  const batch = calls.find((call) => call.path === '/study/reviews');
  return (batch?.body as { reviews: Record<string, unknown>[] } | undefined)?.reviews ?? [];
}

export const start = async () => {
  renderApp('/study');
  fireEvent.click(await screen.findByRole('button', { name: 'Começar' }));
};
