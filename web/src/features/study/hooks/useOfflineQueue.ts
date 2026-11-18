import { useMutation } from '@tanstack/react-query';
import { useCallback, useEffect, useState, useSyncExternalStore } from 'react';
import { submitReviewsMutation } from '../../../api/generated/@tanstack/react-query.gen';
import { pendingReviews } from '../offline/queue';
import { flushPending } from '../offline/sync';
import { csrfHeaders } from './csrf';

const RETRY_MS = 30_000;

function subscribe(onChange: () => void) {
  window.addEventListener('online', onChange);
  window.addEventListener('offline', onChange);
  return () => {
    window.removeEventListener('online', onChange);
    window.removeEventListener('offline', onChange);
  };
}

export const useOnline = () =>
  useSyncExternalStore(
    subscribe,
    () => navigator.onLine,
    () => true,
  );

// A fila de respostas do aparelho: quantas esperam envio, e o envio em si. Reenvia sozinha ao voltar a conexão
// e, enquanto sobrar algo, de tempos em tempos (a rede pode estar "ligada" e o servidor fora).
export function useOfflineQueue() {
  const online = useOnline();
  const [pending, setPending] = useState(0);
  const { mutateAsync } = useMutation(submitReviewsMutation());

  const refresh = useCallback(async () => {
    setPending((await pendingReviews().catch(() => [])).length);
  }, []);

  const flush = useCallback(async () => {
    const sent = await flushPending(
      async (batch) => (await mutateAsync({ headers: csrfHeaders, body: { reviews: [...batch] } })).results,
    );
    await refresh();
    return sent;
  }, [mutateAsync, refresh]);

  useEffect(() => {
    if (online) void flush();
    else void refresh();
  }, [online, flush, refresh]);

  useEffect(() => {
    if (!online || pending === 0) return;
    const timer = setInterval(() => void flush(), RETRY_MS);
    return () => clearInterval(timer);
  }, [online, pending, flush]);

  return { online, pending, flush, refresh };
}
