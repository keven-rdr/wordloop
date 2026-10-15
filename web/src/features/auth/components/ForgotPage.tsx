import * as stylex from '@stylexjs/stylex';
import { Link, Navigate } from '@tanstack/react-router';
import { useTranslation } from 'react-i18next';
import { useSession } from '../hooks/useSession';
import { AuthShell } from './AuthShell';
import { EmailRequestForm } from './EmailRequestForm';
import { form } from './formStyles';
import { ui } from './styles';

export function ForgotPage() {
  const { t, ready } = useTranslation('auth');
  const session = useSession();
  if (session.status === 'authenticated') return <Navigate to="/" replace />;
  if (!ready) return <AuthShell>{null}</AuthShell>;
  return (
    <AuthShell>
      <h1 {...stylex.props(ui.title)}>{t('forgot.title')}</h1>
      <p {...stylex.props(ui.muted)}>{t('forgot.intro')}</p>
      <EmailRequestForm flow="forgot" />
      <p {...stylex.props(form.footnote)}>
        <Link to="/login" {...stylex.props(ui.link)}>
          {t('forgot.back')}
        </Link>
      </p>
    </AuthShell>
  );
}
