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

export function api(cards: unknown[], reviews: (req: Request) => Response | Promise<Response> = accepted) {
  setStoreForTests(); // cada teste começa com a fila do aparelho vazia
  return stubApi({
    'GET /version': () => json(version),
    'GET /me': () => json(me),
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
