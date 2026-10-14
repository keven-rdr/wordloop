import { useMutation, useQueryClient } from '@tanstack/react-query';
import { getMeQueryKey, listPasskeysQueryKey, logoutMutation } from '../../../api/generated/@tanstack/react-query.gen';
import { csrfHeaders } from './csrf';

export function useSignOut() {
  const queryClient = useQueryClient();
  const mutation = useMutation({
    ...logoutMutation(),
    onSuccess: () => {
      queryClient.removeQueries({ queryKey: listPasskeysQueryKey() });
      return queryClient.invalidateQueries({ queryKey: getMeQueryKey() });
    },
  });
  return {
    signOut: () => mutation.mutate({ headers: csrfHeaders }),
    pending: mutation.isPending,
    failed: mutation.isError,
  };
}
