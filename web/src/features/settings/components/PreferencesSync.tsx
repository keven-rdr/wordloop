import { useEffect } from 'react';
import { show, useSettings } from '../hooks/useSettings';

// Quem entra recebe de volta o idioma e o tema que escolheu, em qualquer aparelho. Não desenha nada.
export function PreferencesSync() {
  const { data } = useSettings();
  useEffect(() => {
    if (data) show(data);
  }, [data]);
  return null;
}
