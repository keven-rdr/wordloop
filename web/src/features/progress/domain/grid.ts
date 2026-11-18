import type { ForecastDay, HeatmapDay } from '../../../api/generated/types.gen';

export const WEEKS = 12;

// Quantas respostas no dia pedem cada tom do heatmap (0 = não estudou). Constantes à parte: calibráveis.
const LEVEL_FROM = [1, 5, 15, 30] as const;

export type Level = 0 | 1 | 2 | 3 | 4;

// Uma célula do heatmap. `date` vazio é preenchimento para alinhar a 1ª semana (segunda a domingo).
export type Cell = Readonly<{ key: string; date?: string; reviews: number; level: Level }>;

export function levelOf(reviews: number): Level {
  return LEVEL_FROM.filter((from) => reviews >= from).length as Level;
}

const DAY_MS = 24 * 60 * 60 * 1000;

const parse = (iso: string) => Date.parse(`${iso}T00:00:00Z`);
const format = (ms: number) => new Date(ms).toISOString().slice(0, 10);

export const addDays = (iso: string, days: number) => format(parse(iso) + days * DAY_MS);

// Posição na semana que começa na segunda (0 = segunda, 6 = domingo).
const weekIndex = (iso: string) => (new Date(parse(iso)).getUTCDay() + 6) % 7;

// As 12 semanas até hoje, dia a dia, em colunas de semana (linha 0 = segunda). Hoje é a última célula com data.
export function buildGrid(today: string, days: readonly HeatmapDay[], weeks = WEEKS): Cell[] {
  const reviewsOf = new Map(days.map((day) => [day.date, day.reviews]));
  const start = addDays(today, -(weeks * 7 - 1));
  const padding = Array.from(
    { length: weekIndex(start) },
    (_, at): Cell => ({ key: `pad-${at}`, reviews: 0, level: 0 }),
  );
  const cells = Array.from({ length: weeks * 7 }, (_, at): Cell => {
    const date = addDays(start, at);
    const reviews = reviewsOf.get(date) ?? 0;
    return { key: date, date, reviews, level: levelOf(reviews) };
  });
  return [...padding, ...cells];
}

export const studiedDays = (days: readonly HeatmapDay[]) => days.filter((day) => day.reviews > 0).length;

export const percent = (rate: number) => Math.round(rate * 100);

// Altura (em %) da barra de um dia da previsão: proporcional ao maior dia, com um mínimo para o que não é zero.
export function barHeight(due: number, forecast: readonly ForecastDay[]): number {
  const max = Math.max(...forecast.map((day) => day.due), 0);
  if (due === 0 || max === 0) return 0;
  return Math.max(8, Math.round((due / max) * 100));
}

export const totalDue = (forecast: readonly ForecastDay[]) => forecast.reduce((sum, day) => sum + day.due, 0);
