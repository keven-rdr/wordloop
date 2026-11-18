import { describe, expect, it } from 'vitest';
import { addDays, barHeight, buildGrid, levelOf, percent, studiedDays, totalDue } from './grid';

describe('levelOf', () => {
  it.each([
    [0, 0],
    [1, 1],
    [4, 1],
    [5, 2],
    [14, 2],
    [15, 3],
    [29, 3],
    [30, 4],
    [200, 4],
  ])('%i respostas → tom %i', (reviews, level) => {
    expect(levelOf(reviews)).toBe(level);
  });
});

describe('addDays', () => {
  it('atravessa mês e ano', () => {
    expect(addDays('2026-10-19', -83)).toBe('2026-07-28');
    expect(addDays('2026-12-30', 3)).toBe('2027-01-02');
    expect(addDays('2026-03-01', -1)).toBe('2026-02-28');
  });
});

describe('buildGrid', () => {
  // 19/10/2026 é segunda-feira; 12 semanas = 84 dias, de 28/07 (terça) a 19/10.
  const today = '2026-10-19';
  const grid = buildGrid(today, [
    { date: '2026-10-19', reviews: 20, correct: 15 },
    { date: '2026-07-28', reviews: 3, correct: 3 },
  ]);

  it('começa na segunda: a 1ª semana é preenchida antes de terça (28/07)', () => {
    expect(grid.slice(0, 1).every((cell) => cell.date === undefined)).toBe(true);
    expect(grid[1]?.date).toBe('2026-07-28');
  });

  it('tem 84 dias com data e termina hoje', () => {
    const dated = grid.filter((cell) => cell.date);
    expect(dated).toHaveLength(84);
    expect(dated.at(-1)?.date).toBe(today);
  });

  it('dá o tom pelo número de respostas e zero para dia sem estudo', () => {
    expect(grid.find((cell) => cell.date === today)).toMatchObject({ reviews: 20, level: 3 });
    expect(grid.find((cell) => cell.date === '2026-07-28')).toMatchObject({ reviews: 3, level: 1 });
    expect(grid.find((cell) => cell.date === '2026-09-01')).toMatchObject({ reviews: 0, level: 0 });
  });

  it('cada dia cai na linha da sua semana (segunda = linha 0): hoje, segunda, abre uma coluna nova', () => {
    const at = grid.findIndex((cell) => cell.date === today);
    expect(at % 7).toBe(0);
  });

  it('chaves únicas, para a lista renderizar', () => {
    expect(new Set(grid.map((cell) => cell.key)).size).toBe(grid.length);
  });
});

describe('números do painel', () => {
  const forecast = [
    { date: '2026-10-19', due: 10 },
    { date: '2026-10-20', due: 0 },
    { date: '2026-10-21', due: 2 },
  ];

  it('altura da barra é proporcional ao maior dia, com mínimo para o que não é zero', () => {
    expect(barHeight(10, forecast)).toBe(100);
    expect(barHeight(0, forecast)).toBe(0);
    expect(barHeight(2, forecast)).toBe(20);
    expect(barHeight(1, [{ date: '2026-10-19', due: 100 }])).toBe(8);
    expect(barHeight(0, [{ date: '2026-10-19', due: 0 }])).toBe(0);
  });

  it('soma e porcentagem', () => {
    expect(totalDue(forecast)).toBe(12);
    expect(percent(0.784)).toBe(78);
    expect(percent(1)).toBe(100);
    expect(
      studiedDays([
        { date: 'a', reviews: 0, correct: 0 },
        { date: 'b', reviews: 2, correct: 1 },
      ]),
    ).toBe(1);
  });
});
