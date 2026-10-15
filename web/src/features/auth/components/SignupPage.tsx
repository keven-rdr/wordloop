import * as stylex from '@stylexjs/stylex';
import { Link, Navigate } from '@tanstack/react-router';
import { useTranslation } from 'react-i18next';
import { useSession } from '../hooks/useSession';
import { AuthShell } from './AuthShell';
import { form } from './formStyles';
import { Divider, GoogleButton } from './ProviderButton';
import { ui } from './styles';

export function SignupPage() {
  const { t, ready } = useTranslation('auth');
  const session = useSession();
  if (session.status === 'authenticated') return <Navigate to="/" replace />;
  if (!ready) return <AuthShell>{null}</AuthShell>;
  return (
    <AuthShell>
      <h1 {...stylex.props(ui.title)}>{t('signup.title')}</h1>
      <p role="status" {...stylex.props(ui.muted)}>
        {t('signup.soon')}
      </p>
      <GoogleButton />
      <Divider label={t('signup.orEmail')} />
      <div {...stylex.props(form.field)}>
        <label htmlFor="signup-email" {...stylex.props(form.label)}>
          {t('signup.email')}
        </label>
        <input
          id="signup-email"
          type="email"
          disabled
          placeholder={t('signup.emailPlaceholder')}
          {...stylex.props(form.input, form.inputDisabled)}
        />
      </div>
      <button type="button" disabled {...stylex.props(ui.button, ui.block)}>
        {t('signup.continue')}
      </button>
      <p {...stylex.props(form.footnote)}>
        {t('signup.haveAccount')}{' '}
        <Link to="/login" {...stylex.props(ui.link)}>
          {t('signup.signIn')}
        </Link>
      </p>
    </AuthShell>
  );
}
