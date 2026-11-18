import { useQueryClient } from '@tanstack/react-query';
import { useNavigate } from '@tanstack/react-router';
import { useState } from 'react';
import type { ReviewInput, StudyPack } from '../../../api/generated/types.gen';
import { closeSession, saveReviews, sessionReviews } from '../offline/queue';
import type { StoredReview } from '../offline/store';
import { buildSummary, packKey, summaryKey } from './summary';

export type Destination = 'summary' | 'plan';

// Guarda as respostas no aparelho, tenta enviar e devolve o que há de cada uma (com ou sem confirmação).
// Se nem o armazenamento funcionar, usa as da memória: a sessão termina do mesmo jeito.
async function settle(sessionId: string, reviews: readonly ReviewInput[], flush: () => Promise<boolean>) {
  try {
    await saveReviews(sessionId, reviews);
    await flush();
    return await sessionReviews(sessionId);
  } catch {
    return reviews.map((review): StoredReview => ({ id: review.id, sessionId, review }));
  }
}

// Termina a sessão: sem rede não trava. O que não foi confirmado fica na fila e segue depois;
// o resumo mostra o que já se sabe e avisa quantas respostas ainda esperam.
export function useFinishSession(pack: StudyPack, flush: () => Promise<boolean>) {
  const queryClient = useQueryClient();
  const navigate = useNavigate();
  const [sending, setSending] = useState(false);

  async function finish(reviews: readonly ReviewInput[], destination: Destination) {
    setSending(true);
    const stored = await settle(pack.session_id, reviews, flush);
    queryClient.setQueryData(summaryKey(pack.session_id), buildSummary(pack.cards, stored));
    queryClient.removeQueries({ queryKey: packKey(pack.session_id) });
    await closeSession(pack.session_id).catch(() => undefined);
    if (destination === 'plan') void navigate({ to: '/study' });
    else void navigate({ to: '/study/summary/$id', params: { id: pack.session_id } });
  }

  return { finish, sending };
}
