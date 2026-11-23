import { describe, expect, it } from 'vitest';
import type { Goal, GoalSuggestion } from '../../../api/generated/types.gen';
import { daysPercent, isLastDay, keepRequest, minutesOf, percentOf, requestOf } from './choice';

const suggestions: GoalSuggestion[] = [
  { preset: 'light', daily_minutes: 5, days_target: 7 },
  { preset: 'steady', daily_minutes: 10, days_target: 7 },
  { preset: 'intense', daily_minutes: 20, days_target: 7 },
];

describe('requestOf', () => {
  it('uma sugestão vira o pedido com o mesmo nome e minutos (a meta vale todos os dias)', () => {
    expect(requestOf('steady', suggestions, { minutes: '' })).toEqual({ preset: 'steady', daily_minutes: 10 });
  });

  it('personalizar usa o que foi digitado e vale custom', () => {
    expect(requestOf('custom', suggestions, { minutes: '25' })).toEqual({ preset: 'custom', daily_minutes: 25 });
  });

  it.each([
    ['vazio', ''],
    ['minutos 0', '0'],
    ['minutos 241', '241'],
    ['decimal', '10.5'],
    ['texto', 'dez'],
    ['só espaços', '  '],
  ])('personalizar inválido (%s) não gera pedido', (_name, minutes) => {
    expect(requestOf('custom', suggestions, { minutes })).toBeUndefined();
  });

  it('os limites valem', () => {
    expect(requestOf('custom', suggestions, { minutes: '1' })).toBeDefined();
    expect(requestOf('custom', suggestions, { minutes: '240' })).toBeDefined();
  });

  it('sugestão que o servidor não mandou não gera pedido', () => {
    expect(requestOf('light', [], { minutes: '5' })).toBeUndefined();
  });
});

describe('campos de personalizar', () => {
  it('lê os minutos por dia', () => {
    expect(minutesOf({ minutes: '30' })).toBe(30);
    expect(minutesOf({ minutes: '' })).toBeUndefined();
  });
});

describe('keepRequest', () => {
  it('repete o mesmo tempo por dia da meta', () => {
    const goal: Goal = {
      id: 'g1',
      period_start: '2026-10-12',
      period_end: '2026-10-18',
      preset: 'intense',
      daily_minutes: 20,
      days_target: 4,
      weekdays: [1, 3, 5, 6],
    };
    expect(keepRequest(goal)).toEqual({ preset: 'intense', daily_minutes: 20 });
  });
});

describe('barra de progresso', () => {
  it('porcentagem entre 0 e 100', () => {
    expect(percentOf(3, 5)).toBe(60);
    expect(percentOf(0, 5)).toBe(0);
    expect(percentOf(7, 5)).toBe(100);
    expect(percentOf(1, 0)).toBe(0);
    expect(percentOf(-2, 5)).toBe(0);
  });

  it('dias do ciclo e último dia', () => {
    const progress = {
      days_studied: 2,
      days_target: 4,
      minutes_studied: 0,
      minutes_goal: 0,
      new_words: 0,
      days_left: 0,
      met: false,
    };
    expect(daysPercent(progress)).toBe(50);
    expect(isLastDay(progress)).toBe(true);
    expect(isLastDay({ ...progress, days_left: 3 })).toBe(false);
  });
});
