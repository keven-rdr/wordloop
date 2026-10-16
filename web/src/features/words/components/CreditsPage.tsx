import * as stylex from '@stylexjs/stylex';
import { Link, Navigate } from '@tanstack/react-router';
import { useTranslation } from 'react-i18next';
import { useCredits } from '../hooks/useCredits';
import { useWho } from '../hooks/useWho';
import { ArrowLeft } from './Icons';
import { LoadFailed } from './LoadFailed';
import { w } from './styles';
import { WordsShell } from './WordsShell';

function CreditList() {
  const { credits, loading, failed, retry } = useCredits();
  if (failed) return <LoadFailed onRetry={() => void retry()} />;
  if (loading) return <div aria-busy="true" />;
  return (
    <ul {...stylex.props(w.list)}>
      {credits.map((c) => (
        <li key={c.code} {...stylex.props(w.creditItem)}>
          <span {...stylex.props(w.rowTitle)}>{c.name}</span>
          <span {...stylex.props(w.count)}>{c.license}</span>
          <p {...stylex.props(w.muted)}>{c.notice}</p>
        </li>
      ))}
    </ul>
  );
}

export function CreditsPage() {
  const { t, ready } = useTranslation('words');
  const who = useWho();
  if (who === 'anonymous') return <Navigate to="/login" replace />;
  if (!ready || who === 'loading') return <div aria-busy="true" />;
  const back = (
    <Link to="/words" aria-label={t('credits.back')} {...stylex.props(w.back)}>
      <ArrowLeft {...stylex.props(w.icon)} />
    </Link>
  );
  return (
    <WordsShell title={t('credits.title')} subtitle={t('credits.subtitle')} top={back}>
      <CreditList />
    </WordsShell>
  );
}
