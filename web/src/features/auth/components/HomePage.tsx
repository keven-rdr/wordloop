import * as stylex from '@stylexjs/stylex';
import { Link, Navigate } from '@tanstack/react-router';
import { useTranslation } from 'react-i18next';
import { BrandScreen } from '../../../shared/ui/BrandScreen';
import { CardsIcon } from '../../../shared/ui/nav/Icons';
import { useSession } from '../hooks/useSession';
import { menu } from './menuStyles';

export function HomePage() {
  const { t, ready } = useTranslation('auth');
  const session = useSession();
  if (!ready || session.status === 'loading') return <div aria-busy="true" />;
  if (session.status === 'anonymous') return <Navigate to="/login" replace />;
  const name = session.me.display_name;
  return (
    <BrandScreen title={name ? t('greeting', { name }) : t('greetingAnonymous')} subtitle={t('home.tagline')}>
      <Link to="/words" {...stylex.props(menu.hero)}>
        <CardsIcon {...stylex.props(menu.heroIcon)} />
        <span {...stylex.props(menu.rowText)}>
          <span {...stylex.props(menu.heroTitle)}>{t('myWords')}</span>
          <span {...stylex.props(menu.heroText)}>{t('home.wordsHint')}</span>
        </span>
      </Link>
    </BrandScreen>
  );
}
