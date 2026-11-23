import * as stylex from '@stylexjs/stylex';
import { useTranslation } from 'react-i18next';
import { useDashboard, useHeatmap } from '../hooks/useProgress';
import { Forecast, MostMissed } from './Forecast';
import { Heatmap } from './Heatmap';
import { NumbersRow, StreakCard } from './StatCards';
import { p } from './styles';

// O painel do Início: o que você ganhou e o que vem pela frente. Sem estudo ainda, só um convite.
export function ProgressPanel() {
  const { t, ready } = useTranslation('progress');
  const dashboard = useDashboard();
  const heatmap = useHeatmap();
  if (!ready || dashboard.isPending || heatmap.isPending) return <div aria-busy="true" />;
  if (dashboard.isError || heatmap.isError) {
    return (
      <p role="status" {...stylex.props(p.error)}>
        {t('failed')}
      </p>
    );
  }
  const data = dashboard.data;
  const neverStudied = data.streak === 0 && data.known_words === 0 && heatmap.data.days.length === 0;
  return (
    <section aria-label={t('title')} {...stylex.props(p.panel)}>
      <h2 {...stylex.props(p.heading, p.full)}>{t('title')}</h2>
      {neverStudied ? (
        <p {...stylex.props(p.muted, p.full)}>{t('empty')}</p>
      ) : (
        <>
          <StreakCard dashboard={data} />
          <NumbersRow dashboard={data} />
          <Heatmap today={data.today} days={heatmap.data.days} />
          <Forecast forecast={data.forecast} />
          <MostMissed items={data.most_missed} />
        </>
      )}
    </section>
  );
}
