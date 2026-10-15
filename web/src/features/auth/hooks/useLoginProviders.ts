import { useQuery } from '@tanstack/react-query';
import { listLoginProvidersOptions } from '../../../api/generated/@tanstack/react-query.gen';

type LoginProviders = { providers: readonly string[]; loading: boolean };

export function useLoginProviders(): LoginProviders {
  const { data, isPending } = useQuery({ ...listLoginProvidersOptions(), retry: false, staleTime: 60_000 });
  const providers = data?.providers ?? [];
  return { providers, loading: isPending };
}
