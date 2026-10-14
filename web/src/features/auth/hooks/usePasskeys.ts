import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { listPasskeysOptions, listPasskeysQueryKey } from '../../../api/generated/@tanstack/react-query.gen';
import { deletePasskey } from '../../../api/generated/sdk.gen';
import { csrfHeaders } from './csrf';

export function usePasskeys() {
  return useQuery({ ...listPasskeysOptions(), retry: false });
}

export type DeleteResult = 'deleted' | 'last' | 'failed';

export function useDeletePasskey() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: async (id: string): Promise<DeleteResult> => {
      const { response } = await deletePasskey({ headers: csrfHeaders, path: { id } });
      if (response?.status === 409) return 'last';
      if (response?.ok) return 'deleted';
      return 'failed';
    },
    onSuccess: (result) =>
      result === 'deleted' ? queryClient.invalidateQueries({ queryKey: listPasskeysQueryKey() }) : undefined,
  });
}
