import * as stylex from '@stylexjs/stylex';
import { Link } from '@tanstack/react-router';
import { useTranslation } from 'react-i18next';
import { useSignOut } from '../hooks/useSignOut';
import { ui } from './styles';

export function SignedInPanel({ name }: Readonly<{ name: string }>) {
  const { t } = useTranslation('auth');
  const { signOut, pending, failed } = useSignOut();
  return (
    <div {...stylex.props(ui.panel)}>
      <p>{name ? t('greeting', { name }) : t('greetingAnonymous')}</p>
      <Link to="/words" {...stylex.props(ui.button)}>
        {t('myWords')}
      </Link>
      <Link to="/passkeys" {...stylex.props(ui.button, ui.secondary)}>
        {t('myPasskeys')}
      </Link>
      <button type="button" onClick={signOut} disabled={pending} {...stylex.props(ui.button)}>
        {t('signOut')}
      </button>
      {failed && (
        <p role="alert" {...stylex.props(ui.error)}>
          {t('signOutError')}
        </p>
      )}
    </div>
  );
}
