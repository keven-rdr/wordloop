import { useMutation, useQueryClient } from '@tanstack/react-query';
import { getMeQueryKey, listPasskeysQueryKey } from '../../../api/generated/@tanstack/react-query.gen';
import { beginPasskeyRegistration, finishPasskeyRegistration } from '../../../api/generated/sdk.gen';
import { createCredential } from '../passkey/browser';
import { ensureOk } from '../passkey/errors';
import { csrfHeaders } from './csrf';

type Input = { invite?: string; name?: string };

export function useRegisterPasskey() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: async ({ invite, name }: Input) => {
      const begin = await beginPasskeyRegistration({ headers: csrfHeaders, body: invite ? { invite } : {} });
      ensureOk(begin.response);
      const credential = await createCredential(begin.data?.options);
      const finish = await finishPasskeyRegistration({ headers: csrfHeaders, body: { credential, name } });
      ensureOk(finish.response);
    },
    onSuccess: () =>
      Promise.all([
        queryClient.invalidateQueries({ queryKey: getMeQueryKey() }),
        queryClient.invalidateQueries({ queryKey: listPasskeysQueryKey() }),
      ]),
  });
}
