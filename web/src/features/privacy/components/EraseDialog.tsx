import * as stylex from '@stylexjs/stylex';
import { useEffect, useRef, useState } from 'react';
import { useTranslation } from 'react-i18next';
import { canConfirm, type EraseFailure, MESSAGE_KEYS } from '../domain/erase';
import { d } from './styles';

type Props = Readonly<{
  erasing: boolean;
  failure?: EraseFailure;
  onConfirm: (typed: string) => void;
  onClose: () => void;
  onRelogin: () => void;
}>;

// O aviso de por que não apagou; com login antigo, também o botão para entrar de novo.
function Outcome({ failure, onRelogin }: Readonly<{ failure?: EraseFailure; onRelogin: () => void }>) {
  const { t } = useTranslation('privacy');
  if (!failure) return null;
  return (
    <>
      <p role="alert" {...stylex.props(d.error)}>
        {t(MESSAGE_KEYS[failure])}
      </p>
      {failure === 'recent' && (
        <button type="button" onClick={onRelogin} {...stylex.props(d.action, d.secondary)}>
          {t('erase.relogin')}
        </button>
      )}
    </>
  );
}

// A confirmação da exclusão: a pessoa digita o e-mail da conta (quem confere é o servidor). O foco começa em
// "Cancelar", Esc fecha, e nada acontece sem o texto digitado.
export function EraseDialog({ erasing, failure, onConfirm, onClose, onRelogin }: Props) {
  const { t } = useTranslation('privacy');
  const [typed, setTyped] = useState('');
  const cancel = useRef<HTMLButtonElement>(null);
  useEffect(() => cancel.current?.focus(), []);
  useEffect(() => {
    const onEscape = (event: KeyboardEvent) => event.key === 'Escape' && !erasing && onClose();
    window.addEventListener('keydown', onEscape);
    return () => window.removeEventListener('keydown', onEscape);
  }, [erasing, onClose]);
  return (
    <div {...stylex.props(d.shade)}>
      <form
        role="alertdialog"
        aria-modal="true"
        aria-labelledby="erase-title"
        onSubmit={(event) => {
          event.preventDefault();
          if (canConfirm(typed) && !erasing) onConfirm(typed);
        }}
        {...stylex.props(d.dialog)}
      >
        <h2 id="erase-title" {...stylex.props(d.title)}>
          {t('erase.dialogTitle')}
        </h2>
        <p {...stylex.props(d.text)}>{t('erase.dialogText')}</p>
        <label {...stylex.props(d.field)}>
          <span {...stylex.props(d.label)}>{t('erase.typeEmail')}</span>
          <input
            type="email"
            autoComplete="off"
            autoCapitalize="none"
            spellCheck={false}
            value={typed}
            onChange={(event) => setTyped(event.target.value)}
            {...stylex.props(d.input)}
          />
        </label>
        <Outcome failure={failure} onRelogin={onRelogin} />
        <button type="submit" disabled={!canConfirm(typed) || erasing} {...stylex.props(d.action, d.destructive)}>
          {erasing ? t('erase.erasing') : t('erase.confirm')}
        </button>
        <button
          ref={cancel}
          type="button"
          disabled={erasing}
          onClick={onClose}
          {...stylex.props(d.action, d.secondary)}
        >
          {t('erase.cancel')}
        </button>
      </form>
    </div>
  );
}
