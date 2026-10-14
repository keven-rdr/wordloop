import { useQuery } from '@tanstack/react-query';
import { getMeOptions } from '../../../api/generated/@tanstack/react-query.gen';
import type { Me } from '../../../api/generated/types.gen';

export type Session = { status: 'loading' } | { status: 'anonymous' } | { status: 'authenticated'; me: Me };

export function useSession(): Session {
  const { data, isPending, isError } = useQuery({ ...getMeOptions(), retry: false, refetchOnWindowFocus: false });
  if (isPending) return { status: 'loading' };
  if (isError || !data) return { status: 'anonymous' };
  return { status: 'authenticated', me: data };
}
