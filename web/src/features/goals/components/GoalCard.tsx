import * as stylex from '@stylexjs/stylex';
import { Link } from '@tanstack/react-router';
import { useTranslation } from 'react-i18next';
import { useCurrentGoal } from '../hooks/useGoals';
import { GoalProgressView } from './GoalProgressView';
import { c, g } from './styles';

// O cartão da meta no Início: sem meta convida a definir; com meta mostra o ritmo; fim de ciclo leva à retrospectiva.
// Se o servidor falhar, não mostra nada: o Início não depende da meta.
export function GoalCard() {
  const { t, ready } = useTranslation('goals');
  const current = useCurrentGoal();
  if (!ready || current.isPending || current.isError) return null;
  const { goal, progress, ended } = current.data;

  if (!goal || !progress) {
    return (
      <section aria-labelledby="goal-card" {...stylex.props(g.card)}>
        <h2 id="goal-card" {...stylex.props(g.big)}>
          {t('card.setTitle')}
        </h2>
        <p {...stylex.props(g.text)}>{t('card.setText')}</p>
        <Link to="/goals" {...stylex.props(c.action, c.block)}>
          {t('card.setAction')}
        </Link>
      </section>
    );
  }
  return (
    <section aria-labelledby="goal-card" {...stylex.props(g.card)}>
      <h2 id="goal-card" {...stylex.props(g.title)}>
        {ended ? t('card.endedTitle') : t('card.title')}
      </h2>
      <GoalProgressView progress={progress} ended={ended} />
      <Link to="/goals" {...stylex.props(c.action, c.secondary, c.block)}>
        {ended ? t('card.endedAction') : t('card.see')}
      </Link>
    </section>
  );
}
