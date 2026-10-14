import { useMutation, useQueryClient } from '@tanstack/react-query';
import { getMeQueryKey } from '../../../api/generated/@tanstack/react-query.gen';
import { loginWithPassword } from '../../../api/generated/sdk.gen';
import { ensurePasswordOk } from '../password/errors';
import { csrfHeaders } from './csrf';

type Credentials = { email: string; password: string };

export function usePasswordSignIn() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: async (body: Credentials) => {
      const { response } = await loginWithPassword({ headers: csrfHeaders, body });
      ensurePasswordOk(response, 'login');
    },
    onSuccess: () => queryClient.invalidateQueries({ queryKey: getMeQueryKey() }),
  });
}
