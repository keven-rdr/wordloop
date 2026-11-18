import type { ReviewResult, StudyCard } from '../../../api/generated/types.gen';

// Chaves do cache do TanStack Query: o pacote e o resumo da sessão vivem só na memória do app.
export const packKey = (sessionId: string) => ['study', 'pack', sessionId] as const;
export const summaryKey = (sessionId: string) => ['study', 'summary', sessionId] as const;

export type Summary = Readonly<{ answered: number; correct: number; newWords: number }>;

// Conta o que o servidor confirmou: respostas aceitas, acertos e palavras novas acertadas.
export function buildSummary(
  cards: readonly StudyCard[],
  reviewedCardIds: ReadonlyMap<string, string>,
  results: readonly ReviewResult[],
): Summary {
  const state = new Map(cards.map((card) => [card.id, card.state]));
  const applied = results.filter((result) => result.status !== 'rejected');
  const right = applied.filter((result) => result.correct);
  const cardIds = right.map((result) => reviewedCardIds.get(result.id) ?? '');
  const learned = new Set(cardIds.filter((id) => state.get(id) === 'new'));
  return { answered: applied.length, correct: right.length, newWords: learned.size };
}
