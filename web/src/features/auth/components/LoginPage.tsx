import * as stylex from '@stylexjs/stylex';
import { Link, Navigate, useSearch } from '@tanstack/react-router';
import { useTranslation } from 'react-i18next';
import { usePasskeySignIn } from '../hooks/usePasskeySignIn';
import { useSession } from '../hooks/useSession';
import { PasskeyError } from '../passkey/browser';
import { passkeyMessageKey } from '../passkey/errors';
import { AuthShell } from './AuthShell';
import { form } from './formStyles';
import { LoginForm } from './LoginForm';
import { Divider, SocialButtons } from './ProviderButton';
import { ui } from './styles';

export function LoginPage() {
  const { ready } = useTranslation('auth');
  const session = useSession();
  if (session.status === 'authenticated') return <Navigate to="/" replace />;
  return (
    <AuthShell>{ready && session.status === 'anonymous' ? <SignInOptions /> : <div aria-busy="true" />}</AuthShell>
  );
}

const failures = {
  failed: 'login.errors.failed',
  two_factor: 'login.errors.twoFactor',
  email: 'login.errors.email',
} as const;

function SignInOptions() {
  const { t } = useTranslation('auth');
  const { error } = useSearch({ from: '/login' });
  return (
    <>
      <h1 {...stylex.props(ui.title)}>{t('login.title')}</h1>
      <p {...stylex.props(ui.muted)}>{t('login.subtitle')}</p>
      {error && (
        <p role="alert" {...stylex.props(ui.error)}>
          {t(failures[error])}
        </p>
      )}
      <LoginForm />
      <Divider label={t('login.or')} />
      <PasskeyButton />
      <SocialButtons />
      <p {...stylex.props(form.footnote)}>
        {t('login.noAccount')}{' '}
        <Link to="/signup" {...stylex.props(ui.link)}>
          {t('login.createAccount')}
        </Link>
      </p>
    </>
  );
}

function PasskeyButton() {
  const { t } = useTranslation('auth');
  const signIn = usePasskeySignIn();
  const cancelled = signIn.error instanceof PasskeyError && signIn.error.kind === 'cancelled';
  return (
    <>
      <button
        type="button"
        onClick={() => signIn.mutate()}
        disabled={signIn.isPending}
        {...stylex.props(ui.button, ui.secondary, ui.block)}
      >
        {signIn.isPending ? t('signingIn') : t('signInPasskey')}
      </button>
      {signIn.isError && (
        <p role={cancelled ? 'status' : 'alert'} {...stylex.props(ui.error)}>
          {t(passkeyMessageKey(signIn.error))}
        </p>
      )}
    </>
  );
}
