import * as stylex from '@stylexjs/stylex';
import { type FormEvent, useState } from 'react';
import { useTranslation } from 'react-i18next';
import { usePasswordSignIn } from '../hooks/usePasswordSignIn';
import { passwordMessageKey } from '../password/errors';
import { PasswordField, TextField } from './Fields';
import { form } from './formStyles';
import { ui } from './styles';

export function LoginForm() {
  const { t } = useTranslation('auth');
  const signIn = usePasswordSignIn();
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const submit = (event: FormEvent) => {
    event.preventDefault();
    signIn.mutate({ email: email.trim(), password });
  };
  return (
    <form onSubmit={submit} {...stylex.props(form.stack)}>
      <TextField
        id="login-email"
        label={t('login.email')}
        value={email}
        onChange={setEmail}
        autoComplete="username"
        placeholder={t('login.emailPlaceholder')}
      />
      <PasswordField
        id="login-password"
        label={t('login.password')}
        value={password}
        onChange={setPassword}
        autoComplete="current-password"
      />
      <button type="submit" disabled={signIn.isPending || !email || !password} {...stylex.props(ui.button, ui.block)}>
        {signIn.isPending ? t('login.submitting') : t('login.submit')}
      </button>
      {signIn.isError && (
        <p role="alert" {...stylex.props(ui.error)}>
          {t(passwordMessageKey(signIn.error))}
        </p>
      )}
    </form>
  );
}
