import * as stylex from '@stylexjs/stylex';
import { Link, Navigate } from '@tanstack/react-router';
import { useTranslation } from 'react-i18next';
import { useUnits } from '../hooks/useUnits';
import { useWho } from '../hooks/useWho';
import { DeckContent } from './DeckContent';
import { ArrowLeft } from './Icons';
import { w } from './styles';
import { WordsShell } from './WordsShell';

export function DeckPage({ unitId }: Readonly<{ unitId: string }>) {
  const { t, ready } = useTranslation('words');
  const who = useWho();
  const { units } = useUnits();
  if (who === 'anonymous') return <Navigate to="/login" replace />;
  if (!ready || who === 'loading') return <div aria-busy="true" />;

  const theme = units.find((u) => u.id === unitId)?.theme_key ?? '';
  const title = theme ? t(`units.theme.${theme}`, { defaultValue: theme }) : t('title');
  const back = (
    <Link to="/words" aria-label={t('deck.back')} {...stylex.props(w.back)}>
      <ArrowLeft {...stylex.props(w.icon)} />
    </Link>
  );
  return (
    <WordsShell title={title} top={back} plain>
      <DeckContent unitId={unitId} />
    </WordsShell>
  );
}
