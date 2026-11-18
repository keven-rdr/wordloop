import * as stylex from '@stylexjs/stylex';
import { useTranslation } from 'react-i18next';
import { CloseIcon } from './Icons';
import { s } from './styles';

const progressOf = (done: number, left: number) => (done + left === 0 ? 0 : Math.round((100 * done) / (done + left)));

type Props = Readonly<{ done: number; left: number; onClose: () => void }>;

// ✕ para sair e a barra de exercícios feitos (cartões errados voltam, então o total pode crescer).
export function SessionTop({ done, left, onClose }: Props) {
  const { t } = useTranslation('study');
  const percent = progressOf(done, left);
  return (
    <header {...stylex.props(s.column, s.top)}>
      <button type="button" aria-label={t('session.close')} onClick={onClose} {...stylex.props(s.close)}>
        <CloseIcon {...stylex.props(s.icon)} />
      </button>
      <div
        role="progressbar"
        aria-label={t('session.progress')}
        aria-valuemin={0}
        aria-valuemax={100}
        aria-valuenow={percent}
        {...stylex.props(s.track)}
      >
        <div {...stylex.props(s.fill(percent))} />
      </div>
    </header>
  );
}
