import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { useNavigate, useSearch } from '@tanstack/react-router';
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

// Abre a meta da semana e volta ao Início (ou a Estudar, se a meta foi pedida antes de estudar), já com a meta no cache.
export function useStartGoal() {
  const queryClient = useQueryClient();
  const navigate = useNavigate();
  const search: { from?: string } = useSearch({ strict: false });
  const mutation = useMutation({
    ...startGoalMutation(),
    networkMode: 'always', // sem rede falha na hora e mostra o aviso, em vez de ficar pausada
    onSuccess: (current) => {
      queryClient.setQueryData(getCurrentGoalQueryKey(), current);
      void navigate({ to: search.from === 'study' ? '/study' : '/' });
    },
  });
  const start = (body: GoalRequest) => mutation.mutate({ headers: csrfHeaders, body });
  return { start, saving: mutation.isPending, failed: mutation.isError };
}
