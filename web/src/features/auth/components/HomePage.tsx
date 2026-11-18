import * as stylex from '@stylexjs/stylex';
import { Link, Navigate } from '@tanstack/react-router';
import type { ReactNode } from 'react';
import { useTranslation } from 'react-i18next';
import { BrandScreen } from '../../../shared/ui/BrandScreen';
import { CardsIcon, StudyIcon } from '../../../shared/ui/nav/Icons';
import { useSession } from '../hooks/useSession';
import { menu } from './menuStyles';

// `children` entra depois dos atalhos: o painel de progresso vem da feature progress (a rota junta as duas).
export function HomePage({ children }: Readonly<{ children?: ReactNode }>) {
  const { t, ready } = useTranslation('auth');
  const session = useSession();
  if (!ready || session.status === 'loading') return <div aria-busy="true" />;
  if (session.status === 'anonymous') return <Navigate to="/login" replace />;
  const name = session.me.display_name;
  return (
    <BrandScreen title={name ? t('greeting', { name }) : t('greetingAnonymous')} subtitle={t('home.tagline')}>
      <Link to="/study" {...stylex.props(menu.hero)}>
        <StudyIcon {...stylex.props(menu.heroIcon)} />
        <span {...stylex.props(menu.rowText)}>
          <span {...stylex.props(menu.heroTitle)}>{t('home.studyNow')}</span>
          <span {...stylex.props(menu.heroText)}>{t('home.studyHint')}</span>
        </span>
      </Link>
      <Link to="/words" {...stylex.props(menu.hero)}>
        <CardsIcon {...stylex.props(menu.heroIcon)} />
        <span {...stylex.props(menu.rowText)}>
          <span {...stylex.props(menu.heroTitle)}>{t('myWords')}</span>
          <span {...stylex.props(menu.heroText)}>{t('home.wordsHint')}</span>
        </span>
      </Link>
      {children}
    </BrandScreen>
  );
}
