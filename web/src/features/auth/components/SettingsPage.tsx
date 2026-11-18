import * as stylex from '@stylexjs/stylex';
import { Link, Navigate } from '@tanstack/react-router';
import { useTranslation } from 'react-i18next';
import { ThemeToggle } from '../../../core/theme/ThemeToggle';
import { VersionBadge } from '../../../core/VersionBadge';
import { BrandScreen } from '../../../shared/ui/BrandScreen';
import { Avatar } from '../../../shared/ui/nav/Avatar';
import { useSession } from '../hooks/useSession';
import { useSignOut } from '../hooks/useSignOut';
import { menu } from './menuStyles';

function Chevron() {
  return (
    <svg
      viewBox="0 0 24 24"
      fill="none"
      stroke="currentColor"
      strokeWidth={2}
      strokeLinecap="round"
      strokeLinejoin="round"
      aria-hidden="true"
      {...stylex.props(menu.chevron)}
    >
      <path d="M9 6l6 6-6 6" />
    </svg>
  );
}

type Row = Readonly<{
  to: '/passkeys' | '/preferences' | '/account/data' | '/privacy' | '/words/credits';
  title: string;
  hint: string;
}>;

// Uma linha do menu das Configurações: título, uma dica embaixo e a seta.
function NavRow({ to, title, hint }: Row) {
  return (
    <Link to={to} {...stylex.props(menu.row)}>
      <span {...stylex.props(menu.rowText)}>
        {title}
        <span {...stylex.props(menu.hint)}>{hint}</span>
      </span>
      <Chevron />
    </Link>
  );
}

export function SettingsPage() {
  const { t, ready } = useTranslation('auth');
  const session = useSession();
  const { signOut, pending, failed } = useSignOut();
  if (!ready || session.status === 'loading') return <div aria-busy="true" />;
  if (session.status === 'anonymous') return <Navigate to="/login" replace />;
  const { display_name: name, picture_url: picture } = session.me;
  const rows: Row[] = [
    { to: '/passkeys', title: t('myPasskeys'), hint: t('settings.passkeysHint') },
    { to: '/preferences', title: t('settings.preferences'), hint: t('settings.preferencesHint') },
    { to: '/account/data', title: t('settings.data'), hint: t('settings.dataHint') },
    { to: '/privacy', title: t('settings.privacy'), hint: t('settings.privacyHint') },
    { to: '/words/credits', title: t('settings.credits'), hint: t('settings.creditsHint') },
  ];
  return (
    <BrandScreen
      title={t('settings.title')}
      subtitle={name || t('greetingAnonymous')}
      lead={<Avatar name={name} pictureUrl={picture} large />}
    >
      {rows.map((row) => (
        <NavRow key={row.to} {...row} />
      ))}
      <div {...stylex.props(menu.row)}>
        <span>{t('settings.theme')}</span>
        <ThemeToggle />
      </div>
      <button type="button" onClick={signOut} disabled={pending} {...stylex.props(menu.row, menu.danger)}>
        {t('signOut')}
      </button>
      {failed && (
        <p role="alert" {...stylex.props(menu.error)}>
          {t('signOutError')}
        </p>
      )}
      <VersionBadge />
    </BrandScreen>
  );
}
