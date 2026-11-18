import * as stylex from '@stylexjs/stylex';
import { useTranslation } from 'react-i18next';
import type { GoalProgress } from '../../../api/generated/types.gen';
import { daysPercent, isLastDay } from '../domain/choice';
import { g } from './styles';

// A barra de dias estudados e o que falta; serve para o cartão do Início e para a tela de metas.
export function GoalProgressView({ progress, ended }: Readonly<{ progress: GoalProgress; ended?: boolean }>) {
  const { t } = useTranslation('goals');
  const { days_studied: done, days_target: target } = progress;
  let remaining = t('progress.left', { count: progress.days_left });
  if (isLastDay(progress)) remaining = t('progress.lastDay');
  return (
    <>
      <p {...stylex.props(g.big)}>{t('progress.days', { done, count: target })}</p>
      <div
        role="progressbar"
        aria-label={t('progress.bar')}
        aria-valuemin={0}
        aria-valuemax={100}
        aria-valuenow={daysPercent(progress)}
        {...stylex.props(g.track)}
      >
        <div {...stylex.props(g.fill(daysPercent(progress)))} />
      </div>
      <p {...stylex.props(g.text)}>
        {t('progress.minutes', { done: progress.minutes_studied, goal: progress.minutes_goal })}
      </p>
      {!ended && <p {...stylex.props(g.text)}>{remaining}</p>}
      {progress.met && <p {...stylex.props(g.good)}>{t('progress.met')}</p>}
    </>
  );
}
