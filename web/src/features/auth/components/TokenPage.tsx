import * as stylex from '@stylexjs/stylex';
import { useTranslation } from 'react-i18next';
import { useFragmentSecret } from '../hooks/useFragmentSecret';
import { AuthShell } from './AuthShell';
import { SetPasswordForm } from './SetPasswordForm';
import { ui } from './styles';

const copy = {
  verify: { title: 'verify.title', explain: 'verify.explain' },
  reset: { title: 'reset.title', explain: 'reset.explain' },
} as const;

function TokenPage({ kind }: Readonly<{ kind: keyof typeof copy }>) {
  const { t, ready } = useTranslation('auth');
  const secret = useFragmentSecret();
  if (!ready || secret === undefined) return <div aria-busy="true" />;
  return (
    <AuthShell>
      <h1 {...stylex.props(ui.title)}>{t(copy[kind].title)}</h1>
      {secret === '' ? (
        <p role="alert" {...stylex.props(ui.error)}>
          {t('link.invalid')}
        </p>
      ) : (
        <>
          <p {...stylex.props(ui.muted)}>{t(copy[kind].explain)}</p>
          <SetPasswordForm token={secret} />
        </>
      )}
    </AuthShell>
  );
}

export const VerifyPage = () => <TokenPage kind="verify" />;
export const ResetPage = () => <TokenPage kind="reset" />;
