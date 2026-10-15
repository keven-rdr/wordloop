import * as stylex from '@stylexjs/stylex';
import { type FormEvent, useState } from 'react';
import { useTranslation } from 'react-i18next';
import { emailMessageKey } from '../email/errors';
import { type EmailFlow, useEmailRequest } from '../hooks/useEmailRequest';
import { TextField } from './Fields';
import { form } from './formStyles';
import { ui } from './styles';

const copy = {
  signup: { sent: 'signup.sent', submit: 'signup.continue', field: 'signup.email' },
  forgot: { sent: 'forgot.sent', submit: 'forgot.submit', field: 'forgot.email' },
} as const;

export function EmailRequestForm({ flow }: Readonly<{ flow: EmailFlow }>) {
  const { t } = useTranslation('auth');
  const request = useEmailRequest(flow);
  const [email, setEmail] = useState('');
  if (request.isSuccess) {
    return (
      <p role="status" {...stylex.props(ui.muted)}>
        {t(copy[flow].sent, { email: request.data })}
      </p>
    );
  }
  const submit = (event: FormEvent) => {
    event.preventDefault();
    request.mutate(email.trim());
  };
  return (
    <form onSubmit={submit} {...stylex.props(form.stack)}>
      <TextField
        id={`${flow}-email`}
        label={t(copy[flow].field)}
        value={email}
        onChange={setEmail}
        autoComplete="email"
        placeholder={t('signup.emailPlaceholder')}
      />
      <button type="submit" disabled={request.isPending || !email} {...stylex.props(ui.button, ui.block)}>
        {t(copy[flow].submit)}
      </button>
      {request.isError && (
        <p role="alert" {...stylex.props(ui.error)}>
          {t(emailMessageKey(request.error))}
        </p>
      )}
    </form>
  );
}
