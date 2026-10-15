import { useQuery } from '@tanstack/react-query';
import { getItemOptions } from '../../../api/generated/@tanstack/react-query.gen';

export function useItemDetail(id: string, enabled: boolean) {
  const { data } = useQuery({ ...getItemOptions({ path: { id } }), enabled, retry: false });
  return data;
}
