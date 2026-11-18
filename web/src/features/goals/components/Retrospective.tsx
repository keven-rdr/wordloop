import * as stylex from '@stylexjs/stylex';
import { useTranslation } from 'react-i18next';
import type { GoalProgress } from '../../../api/generated/types.gen';
import { g } from './styles';

function Stat({ value, label }: Readonly<{ value: number; label: string }>) {
  return (
    <li {...stylex.props(g.stat)}>
      <span {...stylex.props(g.statValue)}>{value}</span>
      <span {...stylex.props(g.statLabel)}>{label}</span>
    </li>
  );
}

// Como foi o ciclo que acabou: dias, minutos e palavras novas. Vale cumprida ou não, com tom de incentivo.
export function Retrospective({ progress }: Readonly<{ progress: GoalProgress }>) {
  const { t } = useTranslation('goals');
  const text = progress.met ? 'retro.met' : 'retro.missed';
  return (
    <section aria-labelledby="goal-retro" {...stylex.props(g.card)}>
      <h2 id="goal-retro" {...stylex.props(g.title)}>
        {t('page.retro')}
      </h2>
      <p {...stylex.props(progress.met ? g.good : g.big)}>
        {t(text, { done: progress.days_studied, count: progress.days_target })}
      </p>
      <ul {...stylex.props(g.stats)}>
        <Stat value={progress.days_studied} label={t('retro.days')} />
        <Stat value={progress.minutes_studied} label={t('retro.minutes')} />
        <Stat value={progress.new_words} label={t('retro.words')} />
      </ul>
    </section>
  );
}
