import { describe, expect, it } from 'vitest';
import type { StudyCard } from '../../../api/generated/types.gen';
import { asksDifficulty, type Checked, sessionReducer, startSession } from './sessionReducer';

const T0 = Date.parse('2026-11-18T10:00:00Z');
const choose: StudyCard = {
  id: 'c1',
  item_id: 'i1',
  direction: 'receptive',
  state: 'new',
  exercise: 'choose',
  term: 'house',
  gloss: 'casa',
  mastery: 0,
  options: [
    { item_id: 'i1', gloss: 'casa' },
    { item_id: 'i2', gloss: 'mãe' },
  ],
};
const typed: StudyCard = { ...choose, id: 'c2', item_id: 'i3', exercise: 'type', term: 'mother', gloss: 'mãe' };

const begin = (...cards: StudyCard[]) => startSession('s1', cards, T0);
const asks = (checked: Checked | undefined) => (checked ? asksDifficulty(checked) : 'sem verificação');

describe('sessionReducer', () => {
  it('acerto: grava a resposta com tempo, sugestão e dificuldade escolhida', () => {
    let state = sessionReducer(begin(choose, typed), { type: 'check', now: T0 + 2000, chosenItemId: 'i1' });
    expect(state.checked).toMatchObject({ outcome: 'right', suggested: 'easy' });
    state = sessionReducer(state, { type: 'grade', now: T0 + 2500, id: 'r1', difficulty: 'medium' });
    expect(state.queue.map((c) => c.id)).toEqual(['c2']);
    expect(state.checked).toBeUndefined();
    expect(state.reviews).toEqual([
      expect.objectContaining({
        id: 'r1',
        card_id: 'c1',
        session_id: 's1',
        exercise: 'choose',
        source: 'session',
        chosen_item_id: 'i1',
        client_correct: true,
        response_ms: 2000,
        difficulty_suggested: 'easy',
        difficulty: 'medium',
      }),
    ]);
    expect(state.shownAt).toBe(T0 + 2500);
  });

  it('erro: o cartão volta na fila e vale o motivo informado', () => {
    let state = sessionReducer(begin(choose), { type: 'check', now: T0 + 1000, chosenItemId: 'i2' });
    expect(asks(state.checked)).toBe('errorKind');
    state = sessionReducer(state, { type: 'grade', now: T0 + 1500, id: 'r1', errorKind: 'no_idea' });
    expect(state.queue.map((c) => c.id)).toEqual(['c1']);
    expect(state.reviews[0]).toMatchObject({ client_correct: false, error_kind: 'no_idea', difficulty: undefined });
  });

  it('erro de digitação não pergunta nada e não repete o cartão', () => {
    let state = sessionReducer(begin(typed), { type: 'check', now: T0 + 5000, answer: 'mothar' });
    expect(state.checked?.outcome).toBe('typo');
    expect(asks(state.checked)).toBeUndefined();
    state = sessionReducer(state, { type: 'grade', now: T0 + 6000, id: 'r1' });
    expect(state.queue).toEqual([]);
    expect(state.reviews[0]).toMatchObject({ answer: 'mothar', client_correct: false });
  });

  it('acerto com dica não pergunta e fica limitado a Difícil', () => {
    let state = sessionReducer(begin(typed), { type: 'hint' });
    state = sessionReducer(state, { type: 'check', now: T0 + 4000, answer: 'mother' });
    expect(asks(state.checked)).toBeUndefined();
    state = sessionReducer(state, { type: 'grade', now: T0 + 4500, id: 'r1' });
    expect(state.reviews[0]).toMatchObject({ hint_used: true, difficulty: 'hard' });
    expect(state.hintUsed).toBe(false);
  });

  it('ignora verificar duas vezes e dar nota sem verificar', () => {
    const first = sessionReducer(begin(choose), { type: 'check', now: T0 + 1, chosenItemId: 'i1' });
    expect(sessionReducer(first, { type: 'check', now: T0 + 9, chosenItemId: 'i2' })).toBe(first);
    const fresh = begin(choose);
    expect(sessionReducer(fresh, { type: 'grade', now: T0, id: 'x', difficulty: 'easy' })).toBe(fresh);
  });

  it('sem cartões, verificar não faz nada', () => {
    const empty = begin();
    expect(sessionReducer(empty, { type: 'check', now: T0 })).toBe(empty);
  });
});
