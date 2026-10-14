import * as stylex from '@stylexjs/stylex';
import { useNavigate } from '@tanstack/react-router';
import { useEffect, useState } from 'react';
import { useTranslation } from 'react-i18next';
import { useRegisterPasskey } from '../hooks/useRegisterPasskey';
import { passkeyMessageKey } from '../passkey/errors';
import { AuthShell } from './AuthShell';
import { ui } from './styles';

export function InvitePage() {
  const { t, ready } = useTranslation('auth');
  const navigate = useNavigate();
  const register = useRegisterPasskey();
  const invite = useInviteFromFragment();
  if (!ready || invite === undefined) return <div aria-busy="true" />;
  if (invite === '') {
    return (
      <AuthShell>
        <h1 {...stylex.props(ui.title)}>{t('invite.title')}</h1>
        <p role="alert" {...stylex.props(ui.error)}>
          {t('invite.invalid')}
        </p>
      </AuthShell>
    );
  }
  const create = () => register.mutate({ invite }, { onSuccess: () => void navigate({ to: '/' }) });
  return (
    <AuthShell>
      <h1 {...stylex.props(ui.title)}>{t('invite.title')}</h1>
      <p {...stylex.props(ui.muted)}>{t('invite.explain')}</p>
      <button type="button" onClick={create} disabled={register.isPending} {...stylex.props(ui.button, ui.block)}>
        {register.isPending ? t('invite.creating') : t('invite.create')}
      </button>
      {register.isError && (
        <p role="alert" {...stylex.props(ui.error)}>
          {t(passkeyMessageKey(register.error))}
        </p>
      )}
    </AuthShell>
  );
}

function useInviteFromFragment(): string | undefined {
  const [invite, setInvite] = useState<string | undefined>(undefined);
  useEffect(() => {
    const secret = window.location.hash.slice(1);
    if (secret) {
      setInvite(secret);
      window.history.replaceState(null, '', window.location.pathname);
    } else {
      setInvite((previous) => previous ?? '');
    }
  }, []);
  return invite;
}
