import type { StudyCard } from '../../../api/generated/types.gen';
import type { StoredReview } from '../offline/store';

// Chaves do cache do TanStack Query: o pacote e o resumo da sessão vivem só na memória do app.
export const packKey = (sessionId: string) => ['study', 'pack', sessionId] as const;
export const summaryKey = (sessionId: string) => ['study', 'summary', sessionId] as const;

export type Summary = Readonly<{ answered: number; correct: number; newWords: number; pending: number }>;

// Conta o que o servidor confirmou. Resposta ainda sem confirmação (sem rede) entra pelo palpite do app
// (`client_correct`) e é contada em `pending`: o resumo é provisório para elas.
export function buildSummary(cards: readonly StudyCard[], stored: readonly StoredReview[]): Summary {
  const state = new Map(cards.map((card) => [card.id, card.state]));
  const applied = stored.filter((item) => item.result?.status !== 'rejected');
  const right = applied.filter((item) => Boolean(item.result ? item.result.correct : item.review.client_correct));
  const learned = new Set(right.map((item) => item.review.card_id).filter((id) => state.get(id) === 'new'));
  return {
    answered: applied.length,
    correct: right.length,
    newWords: learned.size,
    pending: stored.filter((item) => !item.result).length,
  };
}
