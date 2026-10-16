import { useQuery } from '@tanstack/react-query';
import { listCreditsOptions } from '../../../api/generated/@tanstack/react-query.gen';

export function useCredits() {
  const { data, isPending, isError, refetch } = useQuery({ ...listCreditsOptions(), retry: false });
  return { credits: data ?? [], loading: isPending, failed: isError, retry: refetch };
}
