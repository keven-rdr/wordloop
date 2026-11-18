import { beforeEach, describe, expect, it } from 'vitest';
import type { ReviewInput, ReviewResult, StudyCard, StudyPack } from '../../../api/generated/types.gen';
import { buildSummary } from '../hooks/summary';
import {
  closeSession,
  collect,
  loadPack,
  pendingReviews,
  recordResults,
  savePack,
  saveReviews,
  sessionReviews,
} from './queue';
import { loadSession, resumeSession } from './resume';
import { memoryStore, setStoreForTests, store } from './store';
import { BATCH_SIZE, flushPending } from './sync';

const review = (id: string, over: Partial<ReviewInput> = {}): ReviewInput => ({
  id,
  card_id: `card-${id}`,
  session_id: 's1',
  exercise: 'choose',
  source: 'session',
  shown_at: '2026-11-18T10:00:00Z',
  answered_at: `2026-11-18T10:00:${id.padStart(2, '0').slice(-2)}Z`,
  client_correct: true,
  ...over,
});
const applied = (id: string, correct = true): ReviewResult => ({ id, status: 'applied', correct });
const card = (id: string, state: StudyCard['state'] = 'new'): StudyCard => ({
  id: `card-${id}`,
  item_id: `item-${id}`,
  direction: 'receptive',
  state,
  exercise: 'choose',
  term: 'house',
  gloss: 'casa',
  mastery: 0,
});
const packOf = (cards: StudyCard[], expiresAt = '2999-01-01T00:00:00Z'): StudyPack => ({
  session_id: 's1',
  cards,
  new_limit: 8,
  review_limit: 20,
  expires_at: expiresAt,
});

beforeEach(() => setStoreForTests(memoryStore()));

describe('fila de respostas', () => {
  it('guarda as respostas e não apaga o resultado de uma que já foi confirmada', async () => {
    await saveReviews('s1', [review('1'), review('2')]);
    await recordResults([applied('1')]);
    await saveReviews('s1', [review('1'), review('2'), review('3')]);

    expect((await pendingReviews()).map((item) => item.id).sort()).toEqual(['2', '3']);
    expect((await sessionReviews('s1')).find((item) => item.id === '1')?.result?.status).toBe('applied');
  });

  it('rejeitada também sai da fila: reenviar não adiantaria', async () => {
    await saveReviews('s1', [review('1')]);
    await recordResults([{ id: '1', status: 'rejected', reason: 'invalid_answer' }]);
    expect(await pendingReviews()).toHaveLength(0);
  });

  it('ao encerrar a sessão, apaga o pacote e o que foi confirmado, mas nunca o que está pendente', async () => {
    await savePack(packOf([card('1'), card('2')]));
    await saveReviews('s1', [review('1'), review('2')]);
    await recordResults([applied('1')]);

    await closeSession('s1');

    expect(await loadPack('s1')).toBeUndefined();
    expect((await sessionReviews('s1')).map((item) => item.id)).toEqual(['2']);
  });

  it('pacote vencido é apagado na limpeza, e a resposta pendente fica', async () => {
    await savePack(packOf([card('1')], '2020-01-01T00:00:00Z'));
    await saveReviews('s1', [review('1')]);
    await collect();
    expect(await loadPack('s1')).toBeUndefined();
    expect(await pendingReviews()).toHaveLength(1);
  });
});

describe('envio', () => {
  it('manda em lotes de até 200 e deixa a fila vazia', async () => {
    const reviews = Array.from({ length: BATCH_SIZE + 5 }, (_, at) => review(String(at)));
    await saveReviews('s1', reviews);
    const sizes: number[] = [];
    const ok = await flushPending(async (batch) => {
      sizes.push(batch.length);
      return batch.map((item) => applied(item.id));
    });
    expect(ok).toBe(true);
    expect(sizes).toEqual([BATCH_SIZE, 5]);
    expect(await pendingReviews()).toHaveLength(0);
  });

  it('se o envio falhar, mantém tudo na fila e devolve false', async () => {
    await saveReviews('s1', [review('1')]);
    const ok = await flushPending(async () => {
      throw new TypeError('Failed to fetch');
    });
    expect(ok).toBe(false);
    expect(await pendingReviews()).toHaveLength(1);
  });

  it('dois pedidos de envio ao mesmo tempo viram um só', async () => {
    await saveReviews('s1', [review('1')]);
    let calls = 0;
    const send = async (batch: readonly ReviewInput[]) => {
      calls += 1;
      return batch.map((item) => applied(item.id));
    };
    await Promise.all([flushPending(send), flushPending(send)]);
    expect(calls).toBe(1);
  });
});

describe('retomar a sessão', () => {
  const cards = [card('1'), card('2'), card('3')];

  it('tira da fila quem já foi acertado; quem errou continua', () => {
    const state = resumeSession(
      { pack: packOf(cards), reviews: [review('1'), review('2', { client_correct: false })] },
      Date.now(),
    );
    expect(state.queue.map((item) => item.id)).toEqual(['card-2', 'card-3']);
    expect(state.reviews).toHaveLength(2);
  });

  it('acha o pacote guardado no aparelho e as respostas em ordem', async () => {
    await savePack(packOf(cards));
    await saveReviews('s1', [review('2'), review('1')]);
    const resume = await loadSession('s1');
    expect(resume?.reviews.map((item) => item.id)).toEqual(['1', '2']);
  });

  it('prefere o pacote da memória', async () => {
    const resume = await loadSession('s1', packOf(cards));
    expect(resume?.pack.cards).toHaveLength(3);
  });

  it.each([
    ['sem pacote em lugar nenhum', undefined],
    ['pacote vencido', packOf(cards, '2020-01-01T00:00:00Z')],
  ])('não retoma: %s', async (_name, cached) => {
    expect(await loadSession('s1', cached)).toBeUndefined();
  });

  it('não retoma sessão já encerrada, mesmo com o pacote na memória', async () => {
    await (await store()).putPack({ sessionId: 's1', pack: packOf(cards), closed: true });
    expect(await loadSession('s1', packOf(cards))).toBeUndefined();
  });
});

describe('resumo com respostas ainda não confirmadas', () => {
  it('usa o servidor onde há resultado, o palpite do app onde não há, e conta as pendentes', () => {
    const summary = buildSummary(
      [card('1'), card('2'), card('3', 'review'), card('4')],
      [
        { id: '1', sessionId: 's1', review: review('1'), result: applied('1', true) },
        { id: '2', sessionId: 's1', review: review('2'), result: applied('2', false) },
        { id: '3', sessionId: 's1', review: review('3') },
        { id: '4', sessionId: 's1', review: review('4'), result: { id: '4', status: 'rejected' } },
      ],
    );
    expect(summary).toEqual({ answered: 3, correct: 2, newWords: 1, pending: 1 });
  });
});
