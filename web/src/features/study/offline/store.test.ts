import 'fake-indexeddb/auto';
import { describe, expect, it } from 'vitest';
import type { ReviewInput, StudyPack } from '../../../api/generated/types.gen';
import { indexedDbStore, memoryStore, type Store } from './store';

const review: ReviewInput = {
  id: 'r1',
  card_id: 'c1',
  exercise: 'type',
  shown_at: '2026-11-18T10:00:00Z',
  answered_at: '2026-11-18T10:00:05Z',
};
const pack: StudyPack = {
  session_id: 's1',
  cards: [],
  new_limit: 8,
  review_limit: 20,
  expires_at: '2999-01-01T00:00:00Z',
};

// O mesmo contrato vale para o IndexedDB e para a memória (a alternativa quando ele não existe).
describe.each([
  ['IndexedDB', indexedDbStore],
  ['memória', async () => memoryStore()],
])('armazenamento em %s', (_name, open: () => Promise<Store>) => {
  it('guarda, lista, atualiza e apaga respostas', async () => {
    const store = await open();
    await store.putReviews([{ id: 'r1', sessionId: 's1', review }]);
    await store.putReviews([
      { id: 'r1', sessionId: 's1', review, result: { id: 'r1', status: 'applied', correct: true } },
    ]);
    const [only, ...rest] = await store.reviews();
    expect(rest).toHaveLength(0);
    expect(only?.result?.status).toBe('applied');
    await store.deleteReviews(['r1']);
    expect(await store.reviews()).toHaveLength(0);
  });

  it('guarda, lista e apaga pacotes', async () => {
    const store = await open();
    await store.putPack({ sessionId: 's1', pack, closed: false });
    expect((await store.packs()).map((item) => item.sessionId)).toEqual(['s1']);
    await store.deletePacks(['s1']);
    expect(await store.packs()).toHaveLength(0);
  });
});

describe('IndexedDB', () => {
  it('os dados sobrevivem a abrir o banco de novo (recarregar a página)', async () => {
    await (await indexedDbStore()).putPack({ sessionId: 's9', pack, closed: false });
    const again = await indexedDbStore();
    expect((await again.packs()).some((item) => item.sessionId === 's9')).toBe(true);
  });
});
