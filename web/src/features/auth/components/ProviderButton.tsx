import * as stylex from '@stylexjs/stylex';
import { useTranslation } from 'react-i18next';
import { form } from './formStyles';
import { ui } from './styles';

export function GoogleButton() {
  const { t } = useTranslation('auth');
  return (
    <button type="button" disabled {...stylex.props(ui.button, ui.block, form.provider)}>
      <img src="/brand/google.svg" alt="" {...stylex.props(form.providerIcon)} />
      {t('provider.google')}
      <span {...stylex.props(form.badge)}>{t('provider.soon')}</span>
    </button>
  );
}

export function Divider({ label }: Readonly<{ label: string }>) {
  return <p {...stylex.props(form.divider)}>{label}</p>;
}
