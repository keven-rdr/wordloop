import * as stylex from '@stylexjs/stylex';
import { useQueryClient } from '@tanstack/react-query';
import { Link, Navigate } from '@tanstack/react-router';
import { useTranslation } from 'react-i18next';
import { BrandScreen } from '../../../shared/ui/BrandScreen';
import { type Summary, summaryKey } from '../hooks/summary';
import { p } from './panelStyles';
import { s } from './styles';

function Stat({ value, label }: Readonly<{ value: number; label: string }>) {
  return (
    <li {...stylex.props(p.stat)}>
      <span {...stylex.props(p.statValue)}>{value}</span>
      <span {...stylex.props(p.statLabel)}>{label}</span>
    </li>
  );
}

// O resumo foca no ganho ("+8 palavras conhecidas"), não só no acerto.
export function SummaryPage({ sessionId }: Readonly<{ sessionId: string }>) {
  const { t, ready } = useTranslation('study');
  const summary = useQueryClient().getQueryData<Summary>(summaryKey(sessionId));
  if (!ready) return <div aria-busy="true" />;
  if (!summary) return <Navigate to="/study" replace />;
  return (
    <BrandScreen title={t('summary.title')}>
      <p {...stylex.props(p.gain)}>
        {summary.newWords > 0 ? t('summary.gain', { count: summary.newWords }) : t('summary.noGain')}
      </p>
      <ul {...stylex.props(p.stats)}>
        <Stat value={summary.answered} label={t('summary.answered')} />
        <Stat value={summary.correct} label={t('summary.correct')} />
        <Stat value={summary.newWords} label={t('summary.newWords')} />
      </ul>
      <Link to="/study" {...stylex.props(s.action, s.linkButton)}>
        {t('summary.more')}
      </Link>
      <Link to="/" {...stylex.props(s.action, s.secondary, s.linkButton)}>
        {t('summary.home')}
      </Link>
    </BrandScreen>
  );
}
