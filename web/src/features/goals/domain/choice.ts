import type { Goal, GoalProgress, GoalRequest, GoalSuggestion } from '../../../api/generated/types.gen';

// Limites da meta (os mesmos do servidor, que é quem decide).
export const MIN_MINUTES = 1;
export const MAX_MINUTES = 240;
export const MIN_DAYS = 1;
export const MAX_DAYS = 7;

// A escolha na tela: uma das 3 sugestões ou "personalizar".
export type Choice = GoalSuggestion['preset'] | 'custom';

// O que a pessoa digitou em "personalizar" (campos de texto: podem estar vazios ou errados).
export type Custom = Readonly<{ minutes: string; days: string }>;

const inRange = (text: string, min: number, max: number) => {
  const value = Number(text);
  return text.trim() !== '' && Number.isInteger(value) && value >= min && value <= max ? value : undefined;
};

export const minutesOf = (custom: Custom) => inRange(custom.minutes, MIN_MINUTES, MAX_MINUTES);
export const daysOf = (custom: Custom) => inRange(custom.days, MIN_DAYS, MAX_DAYS);

// O pedido da escolha atual; vazio se "personalizar" tem campo inválido.
export function requestOf(
  choice: Choice,
  suggestions: readonly GoalSuggestion[],
  custom: Custom,
): GoalRequest | undefined {
  if (choice === 'custom') {
    const minutes = minutesOf(custom);
    const days = daysOf(custom);
    return minutes === undefined || days === undefined
      ? undefined
      : { preset: 'custom', daily_minutes: minutes, days_target: days };
  }
  const picked = suggestions.find((suggestion) => suggestion.preset === choice);
  return picked && { preset: picked.preset, daily_minutes: picked.daily_minutes, days_target: picked.days_target };
}

// "Manter a mesma": os mesmos valores da meta que acabou, para um ciclo novo.
export const keepRequest = (goal: Goal): GoalRequest => ({
  preset: goal.preset,
  daily_minutes: goal.daily_minutes,
  days_target: goal.days_target,
  weekdays: goal.weekdays,
});

// Quanto da barra está cheio (0 a 100).
export const percentOf = (done: number, total: number) =>
  total <= 0 ? 0 : Math.max(0, Math.min(100, Math.round((done / total) * 100)));

export const daysPercent = (progress: GoalProgress) => percentOf(progress.days_studied, progress.days_target);

// Quando o ciclo acaba hoje: último dia para cumprir a meta.
export const isLastDay = (progress: GoalProgress) => progress.days_left === 0;
