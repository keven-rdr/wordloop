import * as stylex from '@stylexjs/stylex';
import { onlineManager, useQuery } from '@tanstack/react-query';
import { Link } from '@tanstack/react-router';
import { useTranslation } from 'react-i18next';
import { getCurrentGoalOptions } from '../../../api/generated/@tanstack/react-query.gen';
import type { CurrentGoal } from '../../../api/generated/types.gen';
import { BrandScreen } from '../../../shared/ui/BrandScreen';
import { useOfflineQueue } from '../hooks/useOfflineQueue';
import { MIN_SESSION_MINUTES, sessionMinutes, useStudyPlan } from '../hooks/useStudyPlan';
import { OfflineBanner } from './OfflineBanner';
import { p } from './panelStyles';
import { s } from './styles';

// Sem meta para esta semana (nunca houve ou o domingo passou): antes de estudar, a pessoa escolhe a meta.
function GoalNeeded() {
  const { t } = useTranslation('study');
  return (
    <>
      <p {...stylex.props(p.lead)}>{t('goal.needed')}</p>
      <Link to="/goals" search={{ from: 'study' }} {...stylex.props(s.action)}>
        {t('goal.define')}
      </Link>
    </>
  );
}

type StartProps = Readonly<{ minutes: number; note?: string }>;

// O tempo de hoje (o da meta da semana) e o botão de começar.
function StartToday({ minutes, note }: StartProps) {
  const { t } = useTranslation('study');
  const { start, loading, failed, empty } = useStudyPlan();
  const { online, pending } = useOfflineQueue();
  return (
    <>
      <div {...stylex.props(p.minute, p.minuteChosen)}>
        <span>{t('goal.today')}</span>
        <span>{t('minutes', { count: sessionMinutes(minutes) })}</span>
      </div>
      {note ? (
        <p {...stylex.props(p.lead)}>{note}</p>
      ) : (
        <Link to="/goals" {...stylex.props(p.lead)}>
          {t('goal.change')}
        </Link>
      )}
      <OfflineBanner online={online} pending={pending} />
      <div role="status">
        {failed && <p {...stylex.props(p.error)}>{t('plan.failed')}</p>}
        {empty && <p {...stylex.props(p.lead)}>{t('plan.empty')}</p>}
      </div>
      <button type="button" disabled={loading} onClick={() => start(minutes)} {...stylex.props(s.action)}>
        {loading ? t('plan.loading') : t('plan.start')}
      </button>
    </>
  );
}

// Estudar: o tempo vem da meta da semana (o mesmo de segunda a domingo); o app não pergunta mais todo dia.
// Na segunda-feira (meta da semana acabou) ou sem meta, pede a meta antes. Se não der para ler a meta (sem rede),
// estuda o mínimo em vez de travar.
export function PlanPage() {
  const { t, ready } = useTranslation('study');
  // Sem rede, a consulta falha na hora (sem pausar nem repetir) e a tela cai no mínimo, com o aviso de conexão.
  const current = useQuery({
    ...getCurrentGoalOptions(),
    networkMode: 'always',
    retry: (failures) => onlineManager.isOnline() && failures < 2,
  });
  if (!ready || current.isPending) return <div aria-busy="true" />;
  return (
    <BrandScreen title={t('title')} subtitle={t('lead')}>
      <Today current={current.data /* numa falha de nova leitura, vale a meta já carregada */} />
    </BrandScreen>
  );
}

// O que mostrar hoje: sem conseguir ler a meta, estuda o mínimo; com meta da semana, o tempo dela; senão, pede a meta.
function Today({ current }: Readonly<{ current?: CurrentGoal }>) {
  const { t } = useTranslation('study');
  if (!current) {
    return <StartToday minutes={MIN_SESSION_MINUTES} note={t('goal.failed', { count: MIN_SESSION_MINUTES })} />;
  }
  if (current.goal && !current.ended) return <StartToday minutes={current.goal.daily_minutes} />;
  return <GoalNeeded />;
}
