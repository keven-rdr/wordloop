import { keepPreviousData, useQuery } from '@tanstack/react-query';
import { listItemsOptions } from '../../../api/generated/@tanstack/react-query.gen';

export const MIN_SEARCH = 2;
const RESULTS = 30;

export type Cefr = 'A1' | 'A2' | 'B1' | 'B2' | 'C1' | 'C2';
export const LEVELS: readonly Cefr[] = ['A1', 'A2', 'B1', 'B2', 'C1', 'C2'];
export type SearchFilters = Readonly<{ unit?: string; cefr?: Cefr }>;

export function hasFilters(filters: SearchFilters): boolean {
  return Boolean(filters.unit ?? filters.cefr);
}

export function useWordSearch(text: string, filters: SearchFilters = {}) {
  const q = text.trim();
  const enabled = q.length >= MIN_SEARCH || hasFilters(filters);
  const { data, isFetching, isError } = useQuery({
    ...listItemsOptions({
      query: {
        q: q.length >= MIN_SEARCH ? q : undefined,
        unit: filters.unit,
        cefr: filters.cefr,
        size: RESULTS,
        sort: 'term,asc',
      },
    }),
    enabled,
    retry: false,
    placeholderData: keepPreviousData,
  });
  const items = data?.items ?? [];
  const total = data?.total ?? 0;
  return { active: enabled, items, total, loading: isFetching, failed: isError };
}
