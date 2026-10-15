import { useQuery } from '@tanstack/react-query';
import { listUnitsOptions } from '../../../api/generated/@tanstack/react-query.gen';

export function useUnits() {
  const { data, isPending, isError, refetch } = useQuery({ ...listUnitsOptions(), retry: false });
  const units = data ?? [];
  return { units, loading: isPending, failed: isError, retry: refetch };
}
