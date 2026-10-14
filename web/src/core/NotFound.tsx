import { Link } from '@tanstack/react-router';
import { useTranslation } from 'react-i18next';
import { AppPage } from '../shared/ui/AppPage';

export function NotFound() {
  const { t } = useTranslation('common');
  return (
    <AppPage>
      <h1>{t('notFound.title')}</h1>
      <Link to="/">{t('notFound.back')}</Link>
    </AppPage>
  );
}
