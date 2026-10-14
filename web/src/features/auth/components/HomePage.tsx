import { useTranslation } from 'react-i18next';
import { AuthPanel } from './AuthPanel';

export function HomePage() {
  const { t } = useTranslation('common');
  return (
    <>
      <h1>{t('app.name')}</h1>
      <p>{t('app.tagline')}</p>
      <AuthPanel />
    </>
  );
}
