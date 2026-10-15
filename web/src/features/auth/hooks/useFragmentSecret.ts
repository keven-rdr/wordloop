import { useEffect, useState } from 'react';

export function useFragmentSecret(): string | undefined {
  const [secret, setSecret] = useState<string | undefined>(undefined);
  useEffect(() => {
    const found = window.location.hash.slice(1);
    if (found) {
      setSecret(found);
      window.history.replaceState(null, '', window.location.pathname);
    } else {
      setSecret((previous) => previous ?? '');
    }
  }, []);
  return secret;
}
