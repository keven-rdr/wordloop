import * as stylex from '@stylexjs/stylex';
import { useTranslation } from 'react-i18next';
import { ArrowLeft, ArrowRight, Eye, EyeOff } from './Icons';
import { w } from './styles';

type Props = Readonly<{
  shown: boolean;
  canPrevious: boolean;
  canNext: boolean;
  onPrevious: () => void;
  onNext: () => void;
  onToggle: () => void;
}>;

export function DeckControls({ shown, canPrevious, canNext, onPrevious, onNext, onToggle }: Props) {
  const { t } = useTranslation('words');
  return (
    <div {...stylex.props(w.controls)}>
      <button
        type="button"
        onClick={onPrevious}
        disabled={!canPrevious}
        aria-label={t('deck.previous')}
        {...stylex.props(w.round)}
      >
        <ArrowLeft {...stylex.props(w.icon)} />
      </button>
      <button
        type="button"
        onClick={onToggle}
        aria-label={t(shown ? 'deck.hide' : 'deck.reveal')}
        {...stylex.props(w.reveal)}
      >
        {shown ? <EyeOff {...stylex.props(w.icon)} /> : <Eye {...stylex.props(w.icon)} />}
      </button>
      <button type="button" onClick={onNext} disabled={!canNext} aria-label={t('deck.next')} {...stylex.props(w.round)}>
        <ArrowRight {...stylex.props(w.icon)} />
      </button>
    </div>
  );
}
