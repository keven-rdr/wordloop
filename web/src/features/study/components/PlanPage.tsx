import * as stylex from '@stylexjs/stylex';
import { useState } from 'react';
import { useTranslation } from 'react-i18next';
import { BrandScreen } from '../../../shared/ui/BrandScreen';
import { PLAN_MINUTES, type PlanMinutes, useStudyPlan } from '../hooks/useStudyPlan';
import { p } from './panelStyles';
import { s } from './styles';

// "Quanto tempo você quer estudar hoje?" Um toque escolhe; o servidor monta o pacote de cartões.
export function PlanPage() {
  const { t, ready } = useTranslation('study');
  const [minutes, setMinutes] = useState<PlanMinutes>(5);
  const { start, loading, failed, empty } = useStudyPlan();
  if (!ready) return <div aria-busy="true" />;
  return (
    <BrandScreen title={t('title')} subtitle={t('lead')}>
      <fieldset {...stylex.props(p.minutes)}>
        <legend {...stylex.props(s.srOnly)}>{t('lead')}</legend>
        {PLAN_MINUTES.map((value) => (
          <button
            key={value}
            type="button"
            aria-pressed={minutes === value}
            onClick={() => setMinutes(value)}
            {...stylex.props(p.minute, minutes === value && p.minuteChosen)}
          >
            {t('minutes', { count: value })}
          </button>
        ))}
      </fieldset>
      <div role="status">
        {failed && <p {...stylex.props(p.error)}>{t('plan.failed')}</p>}
        {empty && <p {...stylex.props(p.lead)}>{t('plan.empty')}</p>}
      </div>
      <button type="button" disabled={loading} onClick={() => start(minutes)} {...stylex.props(s.action)}>
        {loading ? t('plan.loading') : t('plan.start')}
      </button>
    </BrandScreen>
  );
}
