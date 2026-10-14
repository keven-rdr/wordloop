import { useMutation, useQueryClient } from '@tanstack/react-query';
import { getMeQueryKey } from '../../../api/generated/@tanstack/react-query.gen';
import { beginPasskeyLogin, finishPasskeyLogin } from '../../../api/generated/sdk.gen';
import { getCredential } from '../passkey/browser';
import { ensureOk } from '../passkey/errors';
import { csrfHeaders } from './csrf';

export function usePasskeySignIn() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: async () => {
      const begin = await beginPasskeyLogin({ headers: csrfHeaders });
      ensureOk(begin.response);
      const credential = await getCredential(begin.data?.options);
      const finish = await finishPasskeyLogin({ headers: csrfHeaders, body: { credential } });
      ensureOk(finish.response);
    },
    onSuccess: () => queryClient.invalidateQueries({ queryKey: getMeQueryKey() }),
  });
}
