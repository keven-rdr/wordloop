import { useMutation, useQueryClient } from '@tanstack/react-query';
import { useNavigate } from '@tanstack/react-router';
import { useState } from 'react';
import { submitReviewsMutation } from '../../../api/generated/@tanstack/react-query.gen';
import type { ReviewInput, ReviewResult, StudyPack } from '../../../api/generated/types.gen';
import { csrfHeaders } from './csrf';
import { buildSummary, summaryKey } from './summary';

const BATCH_SIZE = 200;

export type Destination = 'summary' | 'plan';

// Envia as respostas (em lotes de até 200) e leva para o resumo ou de volta ao plano.
export function useFinishSession(pack: StudyPack) {
  const queryClient = useQueryClient();
  const navigate = useNavigate();
  const [failed, setFailed] = useState(false);
  const mutation = useMutation(submitReviewsMutation());

  async function sendAll(reviews: readonly ReviewInput[]): Promise<ReviewResult[]> {
    const results: ReviewResult[] = [];
    for (let at = 0; at < reviews.length; at += BATCH_SIZE) {
      const batch = reviews.slice(at, at + BATCH_SIZE);
      const answer = await mutation.mutateAsync({ headers: csrfHeaders, body: { reviews: batch } });
      results.push(...answer.results);
    }
    return results;
  }

  async function finish(reviews: readonly ReviewInput[], destination: Destination) {
    setFailed(false);
    try {
      const results = reviews.length > 0 ? await sendAll(reviews) : [];
      const cardOf = new Map(reviews.map((review) => [review.id, review.card_id]));
      queryClient.setQueryData(summaryKey(pack.session_id), buildSummary(pack.cards, cardOf, results));
      if (destination === 'plan') void navigate({ to: '/study' });
      else void navigate({ to: '/study/summary/$id', params: { id: pack.session_id } });
    } catch {
      setFailed(true);
    }
  }

  return { finish, sending: mutation.isPending, failed };
}
