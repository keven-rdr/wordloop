import { useQuery } from '@tanstack/react-query';
import { getMeOptions } from '../../../api/generated/@tanstack/react-query.gen';

export type Who = 'loading' | 'anonymous' | 'signedIn';

export function useWho(): Who {
  const { data, isPending } = useQuery({ ...getMeOptions(), retry: false, refetchOnWindowFocus: false });
  if (isPending) return 'loading';
  return data ? 'signedIn' : 'anonymous';
}
