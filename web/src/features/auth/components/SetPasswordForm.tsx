import * as stylex from '@stylexjs/stylex';
import { Link, Navigate } from '@tanstack/react-router';
import { type FormEvent, useState } from 'react';
import { useTranslation } from 'react-i18next';
import { useSession } from '../hooks/useSession';
import { useSetPassword } from '../hooks/useSetPassword';
import { passwordMessageKey } from '../password/errors';
import { PasswordField } from './Fields';
import { form } from './formStyles';
import { ui } from './styles';

const minLength = 12;

export function SetPasswordForm({ invite }: Readonly<{ invite: string }>) {
  const { t } = useTranslation('auth');
  const session = useSession();
  const setPassword = useSetPassword();
  const [password, setValue] = useState('');
  const [again, setAgain] = useState('');
  const [saved, setSaved] = useState(false);
  const [problem, setProblem] = useState<'password.tooShort' | 'password.mismatch' | undefined>();
  if (session.status === 'authenticated') return <Navigate to="/" replace />;
  if (saved) return <SavedWithoutSession />;
  const submit = (event: FormEvent) => {
    event.preventDefault();
    if ([...password].length < minLength) return setProblem('password.tooShort');
    if (password !== again) return setProblem('password.mismatch');
    setProblem(undefined);
    setPassword.mutate({ invite, password }, { onSuccess: () => setSaved(true) });
  };
  const serverFailure = setPassword.isError ? passwordMessageKey(setPassword.error) : undefined;
  const failure = problem ?? serverFailure;
  return (
    <form onSubmit={submit} {...stylex.props(form.stack)}>
      <PasswordField
        id="invite-password"
        label={t('invite.password')}
        value={password}
        onChange={setValue}
        autoComplete="new-password"
      />
      <PasswordField
        id="invite-password-again"
        label={t('invite.passwordAgain')}
        value={again}
        onChange={setAgain}
        autoComplete="new-password"
      />
      <p {...stylex.props(ui.muted)}>{t('invite.passwordHint')}</p>
      <button type="submit" disabled={setPassword.isPending} {...stylex.props(ui.button, ui.block)}>
        {setPassword.isPending ? t('invite.settingPassword') : t('invite.setPassword')}
      </button>
      {failure && (
        <p role="alert" {...stylex.props(ui.error)}>
          {t(failure)}
        </p>
      )}
    </form>
  );
}

function SavedWithoutSession() {
  const { t } = useTranslation('auth');
  return (
    <>
      <p role="status">{t('invite.savedWithoutSession')}</p>
      <Link to="/login" {...stylex.props(ui.button, ui.block)}>
        {t('invite.goToLogin')}
      </Link>
    </>
  );
}
