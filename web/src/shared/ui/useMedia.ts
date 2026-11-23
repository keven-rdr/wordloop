import { useSyncExternalStore } from 'react';

const supported = () => typeof window !== 'undefined' && typeof window.matchMedia === 'function';

function subscribe(query: string, onChange: () => void) {
  if (!supported()) return () => undefined;
  const list = window.matchMedia(query);
  // Safari < 14 só tem addListener/removeListener.
  if (typeof list.addEventListener !== 'function') {
    list.addListener?.(onChange);
    return () => list.removeListener?.(onChange);
  }
  list.addEventListener('change', onChange);
  return () => list.removeEventListener('change', onChange);
}

// Se a consulta de mídia vale agora (e acompanha a mudança). Sem `matchMedia` (testes, navegador antigo): `false`,
// isto é, o layout mobile.
export function useMedia(query: string): boolean {
  return useSyncExternalStore(
    (onChange) => subscribe(query, onChange),
    () => supported() && window.matchMedia(query).matches,
    () => false,
  );
}
