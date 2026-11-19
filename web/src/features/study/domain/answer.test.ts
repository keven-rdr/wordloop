import { describe, expect, it } from 'vitest';
import type { StudyCard } from '../../../api/generated/types.gen';
import { checkAssembled, checkChoice, checkTyped, normalize, requeue, suggestDifficulty } from './answer';

const card = (id: string): StudyCard => ({
  id,
  item_id: `item-${id}`,
  direction: 'receptive',
  state: 'new',
  exercise: 'choose',
  term: 'house',
  gloss: 'casa',
  mastery: 0,
});

describe('normalize', () => {
  it('ignora maiúsculas, acentos e espaços a mais', () => {
    expect(normalize('  Ação   Rápida ')).toBe('acao rapida');
  });
});

describe('checkTyped', () => {
  it('aceita igual sem diferenciar maiúsculas, acentos nem espaços extras', () => {
    expect(checkTyped('the cat', '  The   CAT ')).toBe('right');
    expect(checkTyped('café', 'cafe')).toBe('right');
  });

  it('em palavra longa aceita um erro de digitação (troca, falta, sobra e inversão)', () => {
    expect(checkTyped('mother', 'mothar')).toBe('typo');
    expect(checkTyped('mother', 'moter')).toBe('typo');
    expect(checkTyped('mother', 'motherr')).toBe('typo');
    expect(checkTyped('mother', 'mohter')).toBe('typo');
  });

  it('em palavra curta, qualquer diferença é erro', () => {
    expect(checkTyped('cat', 'cut')).toBe('wrong');
  });

  it('dois erros ou resposta em branco é erro', () => {
    expect(checkTyped('mother', 'mxthxr')).toBe('wrong');
    expect(checkTyped('mother', 'mo')).toBe('wrong');
    expect(checkTyped('mother', '   ')).toBe('wrong');
  });
});

describe('checkChoice', () => {
  it('acerta quando escolhe o item do cartão', () => {
    expect(checkChoice(card('1'), 'item-1')).toBe('right');
    expect(checkChoice(card('1'), 'item-2')).toBe('wrong');
  });
});

describe('suggestDifficulty', () => {
  it('escolher: até 3 s fácil, até 8 s médio, depois difícil', () => {
    expect(suggestDifficulty(3000, 'choose')).toBe('easy');
    expect(suggestDifficulty(8000, 'choose')).toBe('medium');
    expect(suggestDifficulty(8001, 'choose')).toBe('hard');
  });

  it('digitar: soma 4 s a cada limite', () => {
    expect(suggestDifficulty(7000, 'type')).toBe('easy');
    expect(suggestDifficulty(12000, 'type')).toBe('medium');
    expect(suggestDifficulty(12001, 'type')).toBe('hard');
  });
});

describe('requeue', () => {
  it('o cartão errado volta daqui a 3 cartões', () => {
    const queue = ['2', '3', '4', '5'].map(card);
    expect(requeue(queue, card('1')).map((c) => c.id)).toEqual(['2', '3', '4', '1', '5']);
  });

  it('com fila curta vai para o fim', () => {
    expect(requeue([card('2')], card('1')).map((c) => c.id)).toEqual(['2', '1']);
    expect(requeue([], card('1')).map((c) => c.id)).toEqual(['1']);
  });
});

describe('checkAssembled', () => {
  const expected = "I don't want to go there, Tom.";

  it('aceita as mesmas palavras na mesma ordem, sem diferenciar maiúsculas, acentos nem pontuação das pontas', () => {
    expect(checkAssembled(expected, "I don't want to go there Tom")).toBe('right');
    expect(checkAssembled(expected, "i DON'T want to go there tom")).toBe('right');
    expect(checkAssembled(expected, "  I  don't want to go there, Tom.  ")).toBe('right');
    expect(checkAssembled('Café, por favor!', 'cafe por favor')).toBe('right');
  });

  it('nunca é "quase": peça errada, fora de ordem, faltando, sobrando ou em branco é erro', () => {
    for (const answer of [
      '',
      "I don't want to go Tom there",
      'I dont want to go there Tom',
      "I don't want to go there",
      "I don't want to go there Tom Tom",
    ]) {
      expect(checkAssembled(expected, answer)).toBe('wrong');
    }
  });
});
