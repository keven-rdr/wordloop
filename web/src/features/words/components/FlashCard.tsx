import * as stylex from '@stylexjs/stylex';
import { useTranslation } from 'react-i18next';
import type { ItemSummary } from '../../../api/generated/types.gen';
import { useItemDetail } from '../hooks/useItemDetail';
import { type CardMotion, motion, motionStyle } from './motion';
import { w } from './styles';

type FaceProps = Readonly<{ item: ItemSummary; shown: boolean; onToggle: () => void; motion?: CardMotion }>;

function Back({ item }: Readonly<{ item: ItemSummary }>) {
  const detail = useItemDetail(item.id, true);
  const example = detail?.examples[0];
  return (
    <>
      <span {...stylex.props(w.hint)}>{item.term}</span>
      <span {...stylex.props(w.gloss)}>{item.gloss}</span>
      {example && (
        <>
          <p {...stylex.props(w.example)}>{example.text}</p>
          {example.translation && <p {...stylex.props(w.exampleGloss)}>{example.translation}</p>}
        </>
      )}
    </>
  );
}

function CardFace({ item, shown, onToggle, motion: kind }: FaceProps) {
  const { t } = useTranslation('words');
  return (
    <button type="button" onClick={onToggle} aria-pressed={shown} {...stylex.props(w.card, kind && motionStyle[kind])}>
      {item.status === 'draft' && <span {...stylex.props(w.badge)}>{t('draft')}</span>}
      {item.cefr && <span {...stylex.props(w.level)}>{item.cefr}</span>}
      {shown ? (
        <Back item={item} />
      ) : (
        <>
          {item.image_url && <img src={item.image_url} alt="" {...stylex.props(w.picture)} />}
          <span {...stylex.props(w.term)}>{item.term}</span>
          <span {...stylex.props(w.hint)}>{t('deck.tap')}</span>
        </>
      )}
    </button>
  );
}

type Props = Readonly<{
  item: ItemSummary;
  shown: boolean;
  onToggle: () => void;
  leaving?: Readonly<{ item: ItemSummary; direction: 'next' | 'previous' }>;
}>;

// Pilha de cartões: o novo sobe para a frente e o anterior sai girando por cima.
export function FlashCard({ item, shown, onToggle, leaving }: Props) {
  return (
    <div {...stylex.props(w.stack)}>
      <div aria-hidden="true" {...stylex.props(w.behind, w.behindFar)} />
      <div aria-hidden="true" {...stylex.props(w.behind, w.behindNear)} />
      <CardFace item={item} shown={shown} onToggle={onToggle} motion={leaving ? 'enter' : undefined} />
      {leaving && (
        <div aria-hidden="true" inert {...stylex.props(motion.overlay)}>
          <CardFace
            item={leaving.item}
            shown={false}
            onToggle={() => undefined}
            motion={leaving.direction === 'next' ? 'leave-next' : 'leave-previous'}
          />
        </div>
      )}
    </div>
  );
}
