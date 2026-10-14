import * as stylex from '@stylexjs/stylex';
import { Link, Navigate } from '@tanstack/react-router';
import { useTranslation } from 'react-i18next';
import { usePasskeySignIn } from '../hooks/usePasskeySignIn';
import { useSession } from '../hooks/useSession';
import { PasskeyError } from '../passkey/browser';
import { passkeyMessageKey } from '../passkey/errors';
import { AuthShell } from './AuthShell';
import { form } from './formStyles';
import { Divider, GoogleButton } from './ProviderButton';
import { ui } from './styles';

export function LoginPage() {
  const { ready } = useTranslation('auth');
  const session = useSession();
  if (session.status === 'authenticated') return <Navigate to="/" replace />;
  return <AuthShell>{ready && session.status === 'anonymous' ? <SignInForm /> : <div aria-busy="true" />}</AuthShell>;
}

function SignInForm() {
  const { t } = useTranslation('auth');
  const signIn = usePasskeySignIn();
  const cancelled = signIn.error instanceof PasskeyError && signIn.error.kind === 'cancelled';
  return (
    <>
      <h1 {...stylex.props(ui.title)}>{t('login.title')}</h1>
      <p {...stylex.props(ui.muted)}>{t('login.subtitle')}</p>
      <button
        type="button"
        onClick={() => signIn.mutate()}
        disabled={signIn.isPending}
        {...stylex.props(ui.button, ui.block)}
      >
        {signIn.isPending ? t('signingIn') : t('signInPasskey')}
      </button>
      {signIn.isError && (
        <p role={cancelled ? 'status' : 'alert'} {...stylex.props(ui.error)}>
          {t(passkeyMessageKey(signIn.error))}
        </p>
      )}
      <Divider label={t('login.or')} />
      <GoogleButton />
      <p {...stylex.props(form.footnote)}>
        {t('login.noAccount')}{' '}
        <Link to="/signup" {...stylex.props(ui.link)}>
          {t('login.createAccount')}
        </Link>
      </p>
    </>
  );
}
