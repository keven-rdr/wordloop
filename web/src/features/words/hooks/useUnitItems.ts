import { useQuery } from '@tanstack/react-query';
import { listItemsOptions } from '../../../api/generated/@tanstack/react-query.gen';

const DECK_SIZE = 100;

export function useUnitItems(unit: string) {
  const query = { unit, size: DECK_SIZE, sort: 'term,asc' } as const;
  const { data, isPending, isError, refetch } = useQuery({ ...listItemsOptions({ query }), retry: false });
  const items = data?.items ?? [];
  return { items, loading: isPending, failed: isError, retry: refetch };
}
