import { describe, expect, it } from 'vitest';
import type { Goal, GoalSuggestion } from '../../../api/generated/types.gen';
import { daysOf, daysPercent, isLastDay, keepRequest, minutesOf, percentOf, requestOf } from './choice';

const suggestions: GoalSuggestion[] = [
  { preset: 'light', daily_minutes: 5, days_target: 2 },
  { preset: 'steady', daily_minutes: 10, days_target: 3 },
  { preset: 'intense', daily_minutes: 20, days_target: 4 },
];

describe('requestOf', () => {
  it('uma sugestão vira o pedido com o mesmo nome, minutos e dias', () => {
    expect(requestOf('steady', suggestions, { minutes: '', days: '' })).toEqual({
      preset: 'steady',
      daily_minutes: 10,
      days_target: 3,
    });
  });

  it('personalizar usa o que foi digitado e vale custom', () => {
    expect(requestOf('custom', suggestions, { minutes: '25', days: '6' })).toEqual({
      preset: 'custom',
      daily_minutes: 25,
      days_target: 6,
    });
  });

  it.each([
    ['vazio', '', ''],
    ['minutos 0', '0', '3'],
    ['minutos 241', '241', '3'],
    ['dias 0', '10', '0'],
    ['dias 8', '10', '8'],
    ['decimal', '10.5', '3'],
    ['texto', 'dez', '3'],
    ['só espaços', '  ', '3'],
  ])('personalizar inválido (%s) não gera pedido', (_name, minutes, days) => {
    expect(requestOf('custom', suggestions, { minutes, days })).toBeUndefined();
  });

  it('os limites valem', () => {
    expect(requestOf('custom', suggestions, { minutes: '1', days: '1' })).toBeDefined();
    expect(requestOf('custom', suggestions, { minutes: '240', days: '7' })).toBeDefined();
  });

  it('sugestão que o servidor não mandou não gera pedido', () => {
    expect(requestOf('light', [], { minutes: '5', days: '2' })).toBeUndefined();
  });
});

describe('campos de personalizar', () => {
  it('lê minutos e dias separadamente', () => {
    expect(minutesOf({ minutes: '30', days: 'x' })).toBe(30);
    expect(daysOf({ minutes: 'x', days: '5' })).toBe(5);
    expect(minutesOf({ minutes: '', days: '' })).toBeUndefined();
  });
});

describe('keepRequest', () => {
  it('repete os mesmos valores da meta, inclusive os dias da semana', () => {
    const goal: Goal = {
      id: 'g1',
      period_start: '2026-10-12',
      period_end: '2026-10-18',
      preset: 'intense',
      daily_minutes: 20,
      days_target: 4,
      weekdays: [1, 3, 5, 6],
    };
    expect(keepRequest(goal)).toEqual({ preset: 'intense', daily_minutes: 20, days_target: 4, weekdays: [1, 3, 5, 6] });
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
