import type { Goal, GoalProgress, GoalRequest, GoalSuggestion } from '../../../api/generated/types.gen';

// Limites da meta (os mesmos do servidor, que é quem decide). A meta é o tempo de cada dia, de segunda a domingo.
export const MIN_MINUTES = 1;
export const MAX_MINUTES = 240;

// A escolha na tela: uma das 3 sugestões ou "personalizar".
export type Choice = GoalSuggestion['preset'] | 'custom';

// O que a pessoa digitou em "personalizar" (campo de texto: pode estar vazio ou errado).
export type Custom = Readonly<{ minutes: string }>;

const inRange = (text: string, min: number, max: number) => {
  const value = Number(text);
  return text.trim() !== '' && Number.isInteger(value) && value >= min && value <= max ? value : undefined;
};

export const minutesOf = (custom: Custom) => inRange(custom.minutes, MIN_MINUTES, MAX_MINUTES);

// O pedido da escolha atual; vazio se "personalizar" tem campo inválido. Vale todos os dias da semana.
export function requestOf(
  choice: Choice,
  suggestions: readonly GoalSuggestion[],
  custom: Custom,
): GoalRequest | undefined {
  if (choice === 'custom') {
    const minutes = minutesOf(custom);
    return minutes === undefined ? undefined : { preset: 'custom', daily_minutes: minutes };
  }
  const picked = suggestions.find((suggestion) => suggestion.preset === choice);
  return picked && { preset: picked.preset, daily_minutes: picked.daily_minutes };
}

// "Manter a mesma": o mesmo tempo por dia da meta que acabou, para a semana nova.
export const keepRequest = (goal: Goal): GoalRequest => ({ preset: goal.preset, daily_minutes: goal.daily_minutes });

// Quanto da barra está cheio (0 a 100).
export const percentOf = (done: number, total: number) =>
  total <= 0 ? 0 : Math.max(0, Math.min(100, Math.round((done / total) * 100)));

export const daysPercent = (progress: GoalProgress) => percentOf(progress.days_studied, progress.days_target);

// Quando o ciclo acaba hoje: último dia para cumprir a meta.
export const isLastDay = (progress: GoalProgress) => progress.days_left === 0;
