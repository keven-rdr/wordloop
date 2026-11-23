import { useMutation, useQueryClient } from '@tanstack/react-query';
import { useNavigate } from '@tanstack/react-router';
import { useTranslation } from 'react-i18next';
import { planSessionMutation } from '../../../api/generated/@tanstack/react-query.gen';
import { savePack } from '../offline/queue';
import { csrfHeaders } from './csrf';
import { packKey } from './summary';

// Limites do tempo de uma sessão (os do servidor). A meta aceita a partir de 1 minuto; a sessão mínima é de 5.
export const MIN_SESSION_MINUTES = 5;
export const MAX_SESSION_MINUTES = 240;

// O tempo da sessão de hoje: o da meta da semana, dentro dos limites da sessão.
export const sessionMinutes = (dailyMinutes: number) =>
  Math.min(MAX_SESSION_MINUTES, Math.max(MIN_SESSION_MINUTES, Math.round(dailyMinutes)));

// Pede o pacote ao servidor com o tempo da meta, guarda na memória e abre a sessão.
export function useStudyPlan() {
  const queryClient = useQueryClient();
  const navigate = useNavigate();
  const { i18n } = useTranslation();
  const mutation = useMutation({
    ...planSessionMutation(),
    networkMode: 'always', // sem rede falha na hora (e mostra o aviso) em vez de ficar pausada para sempre
    onSuccess: (pack) => {
      if (pack.cards.length === 0) return;
      queryClient.setQueryData(packKey(pack.session_id), pack);
      void savePack(pack).catch(() => undefined); // no aparelho: a sessão sobrevive a recarregar a página e a ficar sem rede
      void navigate({ to: '/study/session/$id', params: { id: pack.session_id } });
    },
  });
  const start = (dailyMinutes: number) =>
    mutation.mutate({
      headers: csrfHeaders,
      body: {
        minutes: sessionMinutes(dailyMinutes),
        mode: 'mixed',
        tz: Intl.DateTimeFormat().resolvedOptions().timeZone,
        locale: i18n.language === 'en-US' ? 'en-US' : 'pt-BR',
      },
    });
  return { start, loading: mutation.isPending, failed: mutation.isError, empty: mutation.data?.cards.length === 0 };
}
