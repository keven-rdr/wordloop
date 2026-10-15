import * as stylex from '@stylexjs/stylex';
import { useTranslation } from 'react-i18next';
import { w } from './styles';

export function LoadFailed({ onRetry }: Readonly<{ onRetry: () => void }>) {
  const { t } = useTranslation('words');
  return (
    <div>
      <p role="alert" {...stylex.props(w.error)}>
        {t('error.load')}
      </p>
      <button type="button" onClick={onRetry} {...stylex.props(w.retry)}>
        {t('error.retry')}
      </button>
    </div>
  );
}
