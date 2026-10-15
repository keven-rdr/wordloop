import { useQuery } from '@tanstack/react-query';
import { listLoginProvidersOptions } from '../../../api/generated/@tanstack/react-query.gen';

export function useLoginProviders(): readonly string[] {
  const { data } = useQuery({ ...listLoginProvidersOptions(), retry: false, staleTime: 60_000 });
  return data?.providers ?? [];
}
