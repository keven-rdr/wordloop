import { useQuery } from '@tanstack/react-query';
import { listCreditsOptions } from '../../../api/generated/@tanstack/react-query.gen';

export function useCredits() {
  const { data, isPending, isError, refetch } = useQuery({ ...listCreditsOptions(), retry: false });
  const credits = data ?? [];
  return { credits, loading: isPending, failed: isError, retry: refetch };
}
