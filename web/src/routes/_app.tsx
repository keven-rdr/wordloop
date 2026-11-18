import { createFileRoute } from '@tanstack/react-router';
import { AppLayout } from '../core/AppLayout';
import { PreferencesSync } from '../features/settings/components/PreferencesSync';

// A moldura de quem entrou, mais a volta das preferências salvas (idioma e tema) em qualquer aparelho.
export const Route = createFileRoute('/_app')({
  component: () => (
    <>
      <PreferencesSync />
      <AppLayout />
    </>
  ),
});
