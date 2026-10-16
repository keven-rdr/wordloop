import * as stylex from '@stylexjs/stylex';
import { useEffect, useState } from 'react';
import { useTranslation } from 'react-i18next';
import type { ItemSummary } from '../../../api/generated/types.gen';
import { useUnitItems } from '../hooks/useUnitItems';
import { DeckControls } from './DeckControls';
import { FlashCard } from './FlashCard';
import { LoadFailed } from './LoadFailed';
import { CARD_MS, reducedMotion } from './motion';
import { w } from './styles';

export function DeckContent({ unitId }: Readonly<{ unitId: string }>) {
  const { t } = useTranslation('words');
  const { items, loading, failed, retry } = useUnitItems(unitId);
  const [index, setIndex] = useState(0);
  const [shown, setShown] = useState(false);
  const [leaving, setLeaving] = useState<{ item: ItemSummary; direction: 'next' | 'previous' } | undefined>();
  const item = items[index];
  useEffect(() => {
    if (!leaving) return;
    const timer = setTimeout(() => setLeaving(undefined), CARD_MS);
    return () => clearTimeout(timer);
  }, [leaving]);
  if (failed) return <LoadFailed onRetry={() => void retry()} />;
  if (loading) return <div aria-busy="true" />;
  if (!item) return <p {...stylex.props(w.subtitle)}>{t('deck.empty')}</p>;

  const go = (to: number) => {
    if (leaving) return;
    if (!reducedMotion()) setLeaving({ item, direction: to > index ? 'next' : 'previous' });
    setIndex(to);
    setShown(false);
  };
  const toggle = () => setShown(!shown);
  const percent = ((index + 1) / items.length) * 100;
  return (
    <>
      <div>
        <div
          role="progressbar"
          aria-label={t('deck.progress')}
          aria-valuemin={1}
          aria-valuemax={items.length}
          aria-valuenow={index + 1}
          {...stylex.props(w.progress)}
        >
          <div {...stylex.props(w.fill, w.fillWidth(percent))} />
        </div>
        <p aria-live="polite" {...stylex.props(w.counter)}>
          {t('deck.counter', { current: index + 1, total: items.length })}
        </p>
      </div>
      <FlashCard key={item.id} item={item} shown={shown} onToggle={toggle} leaving={leaving} />
      <DeckControls
        shown={shown}
        canPrevious={index > 0}
        canNext={index < items.length - 1}
        onPrevious={() => go(index - 1)}
        onNext={() => go(index + 1)}
        onToggle={toggle}
      />
    </>
  );
}
