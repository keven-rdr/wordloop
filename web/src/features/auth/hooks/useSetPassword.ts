import { useMutation, useQueryClient } from '@tanstack/react-query';
import { getMeQueryKey } from '../../../api/generated/@tanstack/react-query.gen';
import { setPassword } from '../../../api/generated/sdk.gen';
import { ensurePasswordOk } from '../password/errors';
import { csrfHeaders } from './csrf';

type Input = { invite: string; password: string };

export function useSetPassword() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: async (body: Input) => {
      const { response } = await setPassword({ headers: csrfHeaders, body });
      ensurePasswordOk(response, 'set');
    },
    onSuccess: () => queryClient.invalidateQueries({ queryKey: getMeQueryKey() }),
  });
}
