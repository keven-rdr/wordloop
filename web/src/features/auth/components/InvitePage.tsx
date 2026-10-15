import * as stylex from '@stylexjs/stylex';
import { useNavigate } from '@tanstack/react-router';
import { useEffect, useState } from 'react';
import { useTranslation } from 'react-i18next';
import { useRegisterPasskey } from '../hooks/useRegisterPasskey';
import { passkeyMessageKey } from '../passkey/errors';
import { AuthShell } from './AuthShell';
import { Divider } from './ProviderButton';
import { SetPasswordForm } from './SetPasswordForm';
import { ui } from './styles';

export function InvitePage() {
  const { t, ready } = useTranslation('auth');
  const invite = useInviteFromFragment();
  if (!ready || invite === undefined) return <div aria-busy="true" />;
  return (
    <AuthShell>
      <h1 {...stylex.props(ui.title)}>{t('invite.title')}</h1>
      {invite === '' ? (
        <p role="alert" {...stylex.props(ui.error)}>
          {t('invite.invalid')}
        </p>
      ) : (
        <>
          <p {...stylex.props(ui.muted)}>{t('invite.explain')}</p>
          <SetPasswordForm invite={invite} />
          <Divider label={t('invite.or')} />
          <PasskeyChoice invite={invite} />
        </>
      )}
    </AuthShell>
  );
}

function PasskeyChoice({ invite }: Readonly<{ invite: string }>) {
  const { t } = useTranslation('auth');
  const navigate = useNavigate();
  const register = useRegisterPasskey();
  const create = () => register.mutate({ invite }, { onSuccess: () => void navigate({ to: '/' }) });
  return (
    <>
      <button
        type="button"
        onClick={create}
        disabled={register.isPending}
        {...stylex.props(ui.button, ui.secondary, ui.block)}
      >
        {register.isPending ? t('invite.creating') : t('invite.create')}
      </button>
      {register.isError && (
        <p role="alert" {...stylex.props(ui.error)}>
          {t(passkeyMessageKey(register.error))}
        </p>
      )}
    </>
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
