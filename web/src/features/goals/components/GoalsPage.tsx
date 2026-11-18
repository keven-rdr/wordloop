import * as stylex from '@stylexjs/stylex';
import { useTranslation } from 'react-i18next';
import type { CurrentGoal, GoalSuggestions } from '../../../api/generated/types.gen';
import { BrandScreen } from '../../../shared/ui/BrandScreen';
import { useCurrentGoal, useSuggestions } from '../hooks/useGoals';
import { GoalChooser } from './GoalChooser';
import { GoalProgressView } from './GoalProgressView';
import { Retrospective } from './Retrospective';
import { g } from './styles';

// Onde a pessoa está: retrospectiva (ciclo acabou), meta em andamento ou nada ainda.
function Status({ current }: Readonly<{ current: CurrentGoal }>) {
  const { t } = useTranslation('goals');
  const { goal, progress, ended } = current;
  if (!goal || !progress) return null;
  if (ended) return <Retrospective progress={progress} />;
  return (
    <section aria-labelledby="goal-current" {...stylex.props(g.card)}>
      <h2 id="goal-current" {...stylex.props(g.title)}>
        {t('page.current')}
      </h2>
      <GoalProgressView progress={progress} />
    </section>
  );
}

function nextHeading(current: CurrentGoal) {
  if (!current.goal) return 'page.first';
  return current.ended ? 'page.next' : 'page.change';
}

function Content({ current, suggestions }: Readonly<{ current: CurrentGoal; suggestions: GoalSuggestions }>) {
  const { t } = useTranslation('goals');
  return (
    <BrandScreen title={t('title')} subtitle={t('subtitle')}>
      <Status current={current} />
      <section aria-labelledby="goal-next" {...stylex.props(g.section)}>
        <h2 id="goal-next" {...stylex.props(g.heading)}>
          {t(nextHeading(current))}
        </h2>
        <GoalChooser suggestions={suggestions} goal={current.goal} ended={current.ended} />
      </section>
    </BrandScreen>
  );
}

// A tela de metas: a meta atual (ou a retrospectiva, se o ciclo acabou) e a escolha da próxima.
// Responder não é obrigatório: sem escolher nada, a meta anterior continua valendo.
export function GoalsPage() {
  const { t, ready } = useTranslation('goals');
  const current = useCurrentGoal();
  const suggestions = useSuggestions();
  if (!ready || current.isPending || suggestions.isPending) return <div aria-busy="true" />;
  if (current.isError || suggestions.isError) {
    return (
      <BrandScreen title={t('title')}>
        <p role="status" {...stylex.props(g.error)}>
          {t('failed')}
        </p>
      </BrandScreen>
    );
  }
  return <Content current={current.data} suggestions={suggestions.data} />;
}
