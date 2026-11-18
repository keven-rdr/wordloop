import * as stylex from '@stylexjs/stylex';
import { useTranslation } from 'react-i18next';
import { p } from './panelStyles';
import { s } from './styles';

type Props = Readonly<{ failed: boolean; onRetry: () => void }>;

// Enquanto as respostas vão para o servidor (ou se falhar): nada se perde, dá para tentar de novo.
export function SendingPanel({ failed, onRetry }: Props) {
  const { t } = useTranslation('study');
  return (
    <div role="status" {...stylex.props(p.page)}>
      {failed ? (
        <>
          <p {...stylex.props(p.error)}>{t('sending.failed')}</p>
          <button type="button" onClick={onRetry} {...stylex.props(s.action)}>
            {t('sending.retry')}
          </button>
        </>
      ) : (
        <p {...stylex.props(p.lead)}>{t('sending.busy')}</p>
      )}
    </div>
  );
}
