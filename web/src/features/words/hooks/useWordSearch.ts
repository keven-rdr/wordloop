import { keepPreviousData, useQuery } from '@tanstack/react-query';
import { listItemsOptions } from '../../../api/generated/@tanstack/react-query.gen';

export const MIN_SEARCH = 2;
const RESULTS = 30;

export function useWordSearch(text: string) {
  const q = text.trim();
  const enabled = q.length >= MIN_SEARCH;
  const { data, isFetching, isError } = useQuery({
    ...listItemsOptions({ query: { q, size: RESULTS, sort: 'term,asc' } }),
    enabled,
    retry: false,
    placeholderData: keepPreviousData,
  });
  const items = data?.items ?? [];
  const total = data?.total ?? 0;
  return { active: enabled, items, total, loading: isFetching, failed: isError };
}
