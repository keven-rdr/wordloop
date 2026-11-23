import * as stylex from '@stylexjs/stylex';
import { Link, Navigate } from '@tanstack/react-router';
import type { ReactNode } from 'react';
import { useTranslation } from 'react-i18next';
import { BrandScreen } from '../../../shared/ui/BrandScreen';
import { CardsIcon, StudyIcon } from '../../../shared/ui/nav/Icons';
import { useSession } from '../hooks/useSession';
import { home } from './homeStyles';
import { menu } from './menuStyles';

type Props = Readonly<{ goal?: ReactNode; children?: ReactNode }>;

// `goal` entra logo abaixo dos atalhos e `children` (o painel de progresso) depois: vêm das features goals e progress
// (a rota junta as três). No desktop, atalhos e meta formam uma coluna e o painel fica ao lado.
export function HomePage({ goal, children }: Props) {
  const { t, ready } = useTranslation('auth');
  const session = useSession();
  if (!ready || session.status === 'loading') return <div aria-busy="true" />;
  if (session.status === 'anonymous') return <Navigate to="/login" replace />;
  const name = session.me.display_name;
  return (
    <BrandScreen wide title={name ? t('greeting', { name }) : t('greetingAnonymous')} subtitle={t('home.tagline')}>
      <div {...stylex.props(home.grid)}>
        <div {...stylex.props(home.stack)}>
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
          {goal}
        </div>
        <div {...stylex.props(home.main)}>{children}</div>
      </div>
    </BrandScreen>
  );
}
