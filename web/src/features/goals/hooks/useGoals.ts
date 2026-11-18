import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { useNavigate } from '@tanstack/react-router';
import {
  getCurrentGoalOptions,
  getCurrentGoalQueryKey,
  getGoalSuggestionsOptions,
  startGoalMutation,
} from '../../../api/generated/@tanstack/react-query.gen';
import type { GoalRequest } from '../../../api/generated/types.gen';

// O servidor só aceita escrita com este cabeçalho (proteção contra CSRF).
const csrfHeaders = { 'X-Requested-With': 'XMLHttpRequest' } as const;

// A meta do ciclo e o quanto já foi cumprido.
export const useCurrentGoal = () => useQuery(getCurrentGoalOptions());

// As 3 sugestões (leve, constante, intensa).
export const useSuggestions = () => useQuery(getGoalSuggestionsOptions());

// Abre um ciclo novo e volta ao Início, já com a meta nova no cache.
export function useStartGoal() {
  const queryClient = useQueryClient();
  const navigate = useNavigate();
  const mutation = useMutation({
    ...startGoalMutation(),
    networkMode: 'always', // sem rede falha na hora e mostra o aviso, em vez de ficar pausada
    onSuccess: (current) => {
      queryClient.setQueryData(getCurrentGoalQueryKey(), current);
      void navigate({ to: '/' });
    },
  });
  const start = (body: GoalRequest) => mutation.mutate({ headers: csrfHeaders, body });
  return { start, saving: mutation.isPending, failed: mutation.isError };
}
