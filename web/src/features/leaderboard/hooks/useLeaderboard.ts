import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import {
  getLeaderboardOptions,
  getLeaderboardProfileOptions,
  getLeaderboardProfileQueryKey,
  getMyScoreOptions,
} from '../../../api/generated/@tanstack/react-query.gen';
import { saveLeaderboardProfile } from '../../../api/generated/sdk.gen';
import type { LeaderboardProfile } from '../../../api/generated/types.gen';

// O servidor só aceita escrita com este cabeçalho (proteção contra CSRF).
const csrfHeaders = { 'X-Requested-With': 'XMLHttpRequest' } as const;

export type Period = 'week' | 'all';

// O pódio do período: as primeiras posições e a linha de quem pergunta.
export const useBoard = (period: Period) => useQuery(getLeaderboardOptions({ query: { period, limit: 50 } }));

// Pontos da semana e de sempre, nível e dias da meta cumpridos.
export const useScore = () => useQuery(getMyScoreOptions());

// Como a pessoa aparece no ranking (apelido e se participa).
export const useProfile = () => useQuery(getLeaderboardProfileOptions());

// O resultado de salvar: salvo, apelido em uso (409), inválido (400) ou outro problema.
export type SaveResult = 'saved' | 'taken' | 'invalid' | 'failed';

// Salva o apelido e a participação; salvo, o ranking e o perfil são lidos de novo (a pessoa pode ter entrado ou saído).
export function useSaveProfile() {
  const queryClient = useQueryClient();
  const mutation = useMutation({
    networkMode: 'always',
    mutationFn: async (body: LeaderboardProfile): Promise<SaveResult> => {
      const { response } = await saveLeaderboardProfile({ headers: csrfHeaders, body });
      if (response?.ok) return 'saved';
      if (response?.status === 409) return 'taken';
      if (response?.status === 400) return 'invalid';
      return 'failed';
    },
    onSuccess: (result) => {
      if (result !== 'saved') return;
      void queryClient.invalidateQueries({ queryKey: getLeaderboardProfileQueryKey() });
      void queryClient.invalidateQueries({ queryKey: [{ _id: 'getLeaderboard' }] });
    },
  });
  return {
    save: (body: LeaderboardProfile) => mutation.mutate(body),
    saving: mutation.isPending,
    result: mutation.isError ? ('failed' as const) : mutation.data,
  };
}
