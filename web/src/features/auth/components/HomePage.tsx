import { Navigate } from '@tanstack/react-router';
import { useTranslation } from 'react-i18next';
import { AppPage } from '../../../shared/ui/AppPage';
import { useSession } from '../hooks/useSession';
import { SignedInPanel } from './SignedInPanel';

export function HomePage() {
  const { t } = useTranslation('common');
  const { ready } = useTranslation('auth');
  const session = useSession();
  if (!ready || session.status === 'loading') return <div aria-busy="true" />;
  if (session.status === 'anonymous') return <Navigate to="/login" replace />;
  return (
    <AppPage>
      <h1>{t('app.name')}</h1>
      <p>{t('app.tagline')}</p>
      <SignedInPanel name={session.me.display_name} />
    </AppPage>
  );
}
