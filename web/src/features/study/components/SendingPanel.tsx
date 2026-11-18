import * as stylex from '@stylexjs/stylex';
import { useTranslation } from 'react-i18next';
import { p } from './panelStyles';

// Enquanto as respostas são guardadas e enviadas (sem rede, o envio é adiado e esta tela passa rápido).
export function SendingPanel() {
  const { t } = useTranslation('study');
  return (
    <div role="status" {...stylex.props(p.page)}>
      <p {...stylex.props(p.lead)}>{t('sending.busy')}</p>
    </div>
  );
}
