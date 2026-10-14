import * as stylex from '@stylexjs/stylex';
import { Link } from '@tanstack/react-router';
import { useTranslation } from 'react-i18next';
import { usePasskeySignIn } from '../hooks/usePasskeySignIn';
import { useSession } from '../hooks/useSession';
import { useSignOut } from '../hooks/useSignOut';
import { PasskeyError } from '../passkey/browser';
import { passkeyMessageKey } from '../passkey/errors';
import { ui } from './styles';

export function AuthPanel() {
  const { ready } = useTranslation('auth');
  const session = useSession();
  if (!ready || session.status === 'loading') return <div aria-busy="true" />;
  if (session.status === 'anonymous') return <AnonymousPanel />;
  return <AuthenticatedPanel name={session.me.display_name} />;
}

function AnonymousPanel() {
  const { t } = useTranslation('auth');
  const signIn = usePasskeySignIn();
  const cancelled = signIn.error instanceof PasskeyError && signIn.error.kind === 'cancelled';
  return (
    <div {...stylex.props(ui.panel)}>
      <p>{t('greetingAnonymous')}</p>
      <button type="button" onClick={() => signIn.mutate()} disabled={signIn.isPending} {...stylex.props(ui.button)}>
        {signIn.isPending ? t('signingIn') : t('signInPasskey')}
      </button>
      {signIn.isError && (
        <p role={cancelled ? 'status' : 'alert'} {...stylex.props(ui.error)}>
          {t(passkeyMessageKey(signIn.error))}
        </p>
      )}
    </div>
  );
}

function AuthenticatedPanel({ name }: Readonly<{ name: string }>) {
  const { t } = useTranslation('auth');
  const { signOut, pending, failed } = useSignOut();
  return (
    <div {...stylex.props(ui.panel)}>
      <p>{name ? t('greeting', { name }) : t('greetingAnonymous')}</p>
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
