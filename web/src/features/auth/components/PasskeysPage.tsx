import * as stylex from '@stylexjs/stylex';
import { Link } from '@tanstack/react-router';
import { useTranslation } from 'react-i18next';
import type { Passkey } from '../../../api/generated/types.gen';
import { AppPage } from '../../../shared/ui/AppPage';
import { type DeleteResult, useDeletePasskey, usePasskeys } from '../hooks/usePasskeys';
import { useRegisterPasskey } from '../hooks/useRegisterPasskey';
import { useSession } from '../hooks/useSession';
import { passkeyMessageKey } from '../passkey/errors';
import { ui } from './styles';

export function PasskeysPage() {
  const { t, ready } = useTranslation('auth');
  const session = useSession();
  if (!ready || session.status === 'loading') return <div aria-busy="true" />;
  if (session.status === 'anonymous') {
    return (
      <AppPage>
        <div {...stylex.props(ui.panel)}>
          <p>{t('passkeys.signInFirst')}</p>
          <Link to="/" {...stylex.props(ui.button)}>
            {t('passkeys.home')}
          </Link>
        </div>
      </AppPage>
    );
  }
  return <PasskeysList />;
}

function PasskeysList() {
  const { t } = useTranslation('auth');
  const list = usePasskeys();
  const add = useRegisterPasskey();
  return (
    <AppPage>
      <div {...stylex.props(ui.panel)}>
        <h1>{t('passkeys.title')}</h1>
        <p {...stylex.props(ui.muted)}>{t('passkeys.intro')}</p>
        {list.isError && (
          <p role="alert" {...stylex.props(ui.error)}>
            {t('passkeys.loadFailed')}
          </p>
        )}
        {list.data?.length === 0 && <p>{t('passkeys.empty')}</p>}
        <ul {...stylex.props(ui.list)}>
          {list.data?.map((passkey) => (
            <PasskeyItem key={passkey.id} passkey={passkey} />
          ))}
        </ul>
        <button type="button" onClick={() => add.mutate({})} disabled={add.isPending} {...stylex.props(ui.button)}>
          {add.isPending ? t('passkeys.adding') : t('passkeys.add')}
        </button>
        {add.isError && (
          <p role="alert" {...stylex.props(ui.error)}>
            {t(passkeyMessageKey(add.error))}
          </p>
        )}
        <Link to="/settings" {...stylex.props(ui.button, ui.secondary)}>
          {t('passkeys.back')}
        </Link>
      </div>
    </AppPage>
  );
}

const deleteMessage: Record<Exclude<DeleteResult, 'deleted'>, 'passkeys.lastOne' | 'passkeys.deleteFailed'> = {
  last: 'passkeys.lastOne',
  failed: 'passkeys.deleteFailed',
};

function PasskeyItem({ passkey }: Readonly<{ passkey: Passkey }>) {
  const { t, i18n } = useTranslation('auth');
  const remove = useDeletePasskey();
  const date = (iso: string) => new Date(iso).toLocaleDateString(i18n.language);
  return (
    <li {...stylex.props(ui.item)}>
      <strong>{passkey.name}</strong>
      <span {...stylex.props(ui.muted)}>{t('passkeys.created', { date: date(passkey.created_at) })}</span>
      <span {...stylex.props(ui.muted)}>
        {passkey.last_used_at ? t('passkeys.lastUsed', { date: date(passkey.last_used_at) }) : t('passkeys.neverUsed')}
      </span>
      <button
        type="button"
        onClick={() => remove.mutate(passkey.id)}
        disabled={remove.isPending}
        {...stylex.props(ui.button, ui.secondary)}
      >
        {t('passkeys.delete')}
      </button>
      {remove.data && remove.data !== 'deleted' && (
        <p role="alert" {...stylex.props(ui.error)}>
          {t(deleteMessage[remove.data])}
        </p>
      )}
    </li>
  );
}
