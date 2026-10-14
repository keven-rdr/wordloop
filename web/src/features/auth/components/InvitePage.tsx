import * as stylex from '@stylexjs/stylex';
import { useNavigate } from '@tanstack/react-router';
import { useEffect, useState } from 'react';
import { useTranslation } from 'react-i18next';
import { useRegisterPasskey } from '../hooks/useRegisterPasskey';
import { passkeyMessageKey } from '../passkey/errors';
import { ui } from './styles';

export function InvitePage() {
  const { t, ready } = useTranslation('auth');
  const navigate = useNavigate();
  const register = useRegisterPasskey();
  const invite = useInviteFromFragment();
  if (!ready || invite === undefined) return <div aria-busy="true" />;
  if (invite === '') {
    return (
      <div {...stylex.props(ui.panel)}>
        <h1>{t('invite.title')}</h1>
        <p role="alert" {...stylex.props(ui.error)}>
          {t('invite.invalid')}
        </p>
      </div>
    );
  }
  const create = () => register.mutate({ invite }, { onSuccess: () => void navigate({ to: '/' }) });
  return (
    <div {...stylex.props(ui.panel)}>
      <h1>{t('invite.title')}</h1>
      <p>{t('invite.explain')}</p>
      <button type="button" onClick={create} disabled={register.isPending} {...stylex.props(ui.button)}>
        {register.isPending ? t('invite.creating') : t('invite.create')}
      </button>
      {register.isError && (
        <p role="alert" {...stylex.props(ui.error)}>
          {t(passkeyMessageKey(register.error))}
        </p>
      )}
    </div>
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
