import { useMutation, useQueryClient } from '@tanstack/react-query';
import { useNavigate } from '@tanstack/react-router';
import { useTranslation } from 'react-i18next';
import { planSessionMutation } from '../../../api/generated/@tanstack/react-query.gen';
import { savePack } from '../offline/queue';
import { csrfHeaders } from './csrf';
import { packKey } from './summary';

export const PLAN_MINUTES = [5, 15, 30, 60] as const;
export type PlanMinutes = (typeof PLAN_MINUTES)[number];

// "Quanto tempo hoje?": pede o pacote ao servidor, guarda na memória e abre a sessão.
export function useStudyPlan() {
  const queryClient = useQueryClient();
  const navigate = useNavigate();
  const { i18n } = useTranslation();
  const mutation = useMutation({
    ...planSessionMutation(),
    onSuccess: (pack) => {
      if (pack.cards.length === 0) return;
      queryClient.setQueryData(packKey(pack.session_id), pack);
      void savePack(pack).catch(() => undefined); // no aparelho: a sessão sobrevive a recarregar a página e a ficar sem rede
      void navigate({ to: '/study/session/$id', params: { id: pack.session_id } });
    },
  });
  const start = (minutes: PlanMinutes) =>
    mutation.mutate({
      headers: csrfHeaders,
      body: {
        minutes,
        mode: 'mixed',
        tz: Intl.DateTimeFormat().resolvedOptions().timeZone,
        locale: i18n.language === 'en-US' ? 'en-US' : 'pt-BR',
      },
    });
  return { start, loading: mutation.isPending, failed: mutation.isError, empty: mutation.data?.cards.length === 0 };
}
