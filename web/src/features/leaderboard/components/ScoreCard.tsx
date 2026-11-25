import * as stylex from '@stylexjs/stylex';
import { Link } from '@tanstack/react-router';
import { useTranslation } from 'react-i18next';
import type { MyScore } from '../../../api/generated/types.gen';
import { useScore } from '../hooks/useLeaderboard';
import { l } from './styles';

// Quanto do nível atual já foi feito (0 a 100).
export const levelPercent = (s: MyScore) =>
  s.next_level <= s.level_floor
    ? 100
    : Math.max(0, Math.min(100, Math.round(((s.total_points - s.level_floor) / (s.next_level - s.level_floor)) * 100)));

// O cartão de pontos no Início: pontos da semana, nível com a barra e dias da meta cumpridos; leva ao ranking.
// Se o servidor falhar, não mostra nada: o Início não depende do ranking.
export function ScoreCard() {
  const { t, ready } = useTranslation('leaderboard');
  const score = useScore();
  if (!ready || score.isPending || score.isError) return null;
  const s = score.data;
  const missing = Math.max(0, s.next_level - s.total_points);
  return (
    <section aria-labelledby="score-card" {...stylex.props(l.card)}>
      <h2 id="score-card" {...stylex.props(l.title)}>
        {t('score.title')}
      </h2>
      <p {...stylex.props(l.big)}>{t('score.week', { count: s.week_points })}</p>
      <p {...stylex.props(l.text)}>{t('score.levelLine', { level: s.level, count: s.total_points })}</p>
      <div
        role="progressbar"
        aria-label={t('score.bar')}
        aria-valuemin={0}
        aria-valuemax={100}
        aria-valuenow={levelPercent(s)}
        {...stylex.props(l.track)}
      >
        <div {...stylex.props(l.fill(levelPercent(s)))} />
      </div>
      <p {...stylex.props(l.text)}>{t('score.toNext', { count: missing, next: s.level + 1 })}</p>
      <p {...stylex.props(l.text)}>{t('score.days', { count: s.week_days_met })}</p>
      <Link to="/ranking" {...stylex.props(l.action, l.secondary)}>
        {t('score.see')}
      </Link>
    </section>
  );
}
