import { useMutation, useQueryClient } from '@tanstack/react-query';
import { getMeQueryKey, logoutMutation } from '../../../api/generated/@tanstack/react-query.gen';

export function useSignOut() {
  const queryClient = useQueryClient();
  const mutation = useMutation({
    ...logoutMutation(),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: getMeQueryKey() }),
  });
  return {
    signOut: () => mutation.mutate({ headers: { 'X-Requested-With': 'XMLHttpRequest' } }),
    pending: mutation.isPending,
    failed: mutation.isError,
  };
}
