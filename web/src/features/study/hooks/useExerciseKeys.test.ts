import { describe, expect, it } from 'vitest';
import type { StudyCard } from '../../../api/generated/types.gen';
import type { Checked } from './sessionReducer';
import { keyAction } from './useExerciseKeys';

const card: StudyCard = {
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
const checked = (over: Partial<Checked>): Checked => ({
  outcome: 'right',
  hintUsed: false,
  answeredAt: 0,
  suggested: 'medium',
  ...over,
});

describe('keyAction', () => {
  it('A–D e 1–4 escolhem a opção; fora do alcance não faz nada', () => {
    expect(keyAction('b', { card })).toEqual({ do: 'choose', itemId: 'i2' });
    expect(keyAction('1', { card })).toEqual({ do: 'choose', itemId: 'i1' });
    expect(keyAction('d', { card })).toBeUndefined();
    expect(keyAction('x', { card })).toBeUndefined();
  });

  it('Enter verifica só com uma opção escolhida', () => {
    expect(keyAction('Enter', { card })).toBeUndefined();
    expect(keyAction('Enter', { card, chosen: 'i1' })).toEqual({ do: 'verify' });
  });

  it('no digitar as teclas de opção não fazem nada', () => {
    expect(keyAction('a', { card: { ...card, exercise: 'type' } })).toBeUndefined();
  });

  it('acerto: 1 Difícil, 3 Fácil e Enter aceita a sugestão', () => {
    const c = checked({});
    expect(keyAction('1', { card, checked: c })).toEqual({ do: 'grade', grade: { difficulty: 'hard' } });
    expect(keyAction('3', { card, checked: c })).toEqual({ do: 'grade', grade: { difficulty: 'easy' } });
    expect(keyAction('Enter', { card, checked: c })).toEqual({ do: 'grade', grade: { difficulty: 'medium' } });
    expect(keyAction('9', { card, checked: c })).toBeUndefined();
  });

  it('erro: 1, 2 e 3 dizem como foi; Enter não escolhe por você', () => {
    const c = checked({ outcome: 'wrong' });
    expect(keyAction('2', { card, checked: c })).toEqual({ do: 'grade', grade: { errorKind: 'almost' } });
    expect(keyAction('Enter', { card, checked: c })).toBeUndefined();
    expect(keyAction('7', { card, checked: c })).toBeUndefined();
  });

  it('sem pergunta (erro de digitação), Enter continua', () => {
    const c = checked({ outcome: 'typo' });
    expect(keyAction('Enter', { card, checked: c })).toEqual({ do: 'grade', grade: {} });
    expect(keyAction('1', { card, checked: c })).toBeUndefined();
  });

  describe('montar frase', () => {
    const assemble: StudyCard = { ...card, exercise: 'assemble', options: undefined, tiles: ['b', 'a', 'c'] };
    const open = { remaining: 3, complete: false };

    it('1 a 9 colocam a peça daquela posição do banco; fora do alcance não faz nada', () => {
      expect(keyAction('1', { card: assemble, assemble: open })).toEqual({ do: 'place', position: 0 });
      expect(keyAction('3', { card: assemble, assemble: open })).toEqual({ do: 'place', position: 2 });
      expect(keyAction('4', { card: assemble, assemble: open })).toBeUndefined();
      expect(keyAction('0', { card: assemble, assemble: open })).toBeUndefined();
      expect(keyAction('a', { card: assemble, assemble: open })).toBeUndefined(); // A–D são das opções, não das peças
    });

    it('Backspace desfaz e Enter só verifica com a frase completa', () => {
      expect(keyAction('Backspace', { card: assemble, assemble: open })).toEqual({ do: 'undo' });
      expect(keyAction('Enter', { card: assemble, assemble: open })).toBeUndefined();
      expect(keyAction('Enter', { card: assemble, assemble: { remaining: 0, complete: true } })).toEqual({
        do: 'verify',
      });
    });

    it('depois de verificar valem as teclas da dificuldade, como nos outros exercícios', () => {
      expect(keyAction('2', { card: assemble, assemble: open, checked: checked({}) })).toEqual({
        do: 'grade',
        grade: { difficulty: 'medium' },
      });
    });
  });
});
