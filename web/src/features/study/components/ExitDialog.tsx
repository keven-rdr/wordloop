import * as stylex from '@stylexjs/stylex';
import { useEffect, useRef } from 'react';
import { useTranslation } from 'react-i18next';
import { s } from './styles';

type Props = Readonly<{ onStay: () => void; onLeave: () => void }>;

// Sair salva o que já foi respondido; por isso a confirmação é leve e o foco começa em "Continuar".
export function ExitDialog({ onStay, onLeave }: Props) {
  const { t } = useTranslation('study');
  const stay = useRef<HTMLButtonElement>(null);
  useEffect(() => stay.current?.focus(), []);
  return (
    <div {...stylex.props(s.shade)}>
      <div role="alertdialog" aria-modal="true" aria-labelledby="exit-title" {...stylex.props(s.dialog)}>
        <h2 id="exit-title" {...stylex.props(s.dialogTitle)}>
          {t('exit.title')}
        </h2>
        <p {...stylex.props(s.dialogText)}>{t('exit.text')}</p>
        <button ref={stay} type="button" onClick={onStay} {...stylex.props(s.action)}>
          {t('exit.stay')}
        </button>
        <button type="button" onClick={onLeave} {...stylex.props(s.action, s.secondary)}>
          {t('exit.leave')}
        </button>
      </div>
    </div>
  );
}
