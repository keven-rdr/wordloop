import * as stylex from '@stylexjs/stylex';
import { Link, Navigate } from '@tanstack/react-router';
import { useTranslation } from 'react-i18next';
import { useSession } from '../hooks/useSession';
import { AuthShell } from './AuthShell';
import { EmailRequestForm } from './EmailRequestForm';
import { form } from './formStyles';
import { Divider, SocialButtons } from './ProviderButton';
import { ui } from './styles';

export function SignupPage() {
  const { t, ready } = useTranslation('auth');
  const session = useSession();
  if (session.status === 'authenticated') return <Navigate to="/" replace />;
  if (!ready) return <AuthShell>{null}</AuthShell>;
  return (
    <AuthShell>
      <h1 {...stylex.props(ui.title)}>{t('signup.title')}</h1>
      <p {...stylex.props(ui.muted)}>{t('signup.intro')}</p>
      <SocialButtons />
      <Divider label={t('signup.orEmail')} />
      <EmailRequestForm flow="signup" />
      <p {...stylex.props(form.footnote)}>
        {t('signup.haveAccount')}{' '}
        <Link to="/login" {...stylex.props(ui.link)}>
          {t('signup.signIn')}
        </Link>
      </p>
    </AuthShell>
  );
}
