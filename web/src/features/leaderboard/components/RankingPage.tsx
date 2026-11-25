import * as stylex from '@stylexjs/stylex';
import { useState } from 'react';
import { useTranslation } from 'react-i18next';
import { BrandScreen } from '../../../shared/ui/BrandScreen';
import { panel } from '../../../shared/ui/panelStyles';
import { type Period, useBoard, useProfile } from '../hooks/useLeaderboard';
import { JoinRanking } from './JoinRanking';
import { Podium } from './Podium';
import { RankingList } from './RankingList';
import { l } from './styles';

const periods: readonly Period[] = ['week', 'all'];

function Tabs({ period, onChange }: Readonly<{ period: Period; onChange: (next: Period) => void }>) {
  const { t } = useTranslation('leaderboard');
  return (
    <fieldset aria-label={t('tabs.label')} {...stylex.props(l.tabs)}>
      {periods.map((value) => (
        <button
          key={value}
          type="button"
          aria-pressed={period === value}
          onClick={() => onChange(value)}
          {...stylex.props(l.tab, period === value && l.tabOn)}
        >
          {t(`tabs.${value}`)}
        </button>
      ))}
    </fieldset>
  );
}

// O pódio do período escolhido e a lista; sem ninguém pontuando, um convite a estudar.
function Board({ period }: Readonly<{ period: Period }>) {
  const { t } = useTranslation('leaderboard');
  const board = useBoard(period);
  if (board.isPending) return <div aria-busy="true" />;
  if (board.isError) {
    return <output {...stylex.props(panel.status, panel.error)}>{t('failed')}</output>;
  }
  const data = board.data;
  return (
    <section aria-label={t(`tabs.${period}`)} {...stylex.props(panel.section)}>
      {data.entries.length === 0 && <p {...stylex.props(panel.text)}>{t('empty')}</p>}
      {data.entries.length > 0 && <Podium entries={data.entries} />}
      <RankingList board={data} />
    </section>
  );
}

// Ranking: abas Semana e Geral, pódio, lista e a participação (apelido).
export function RankingPage() {
  const { t, ready } = useTranslation('leaderboard');
  const [period, setPeriod] = useState<Period>('week');
  const profile = useProfile();
  if (!ready) return <div aria-busy="true" />;
  return (
    <BrandScreen title={t('title')} subtitle={t('subtitle')}>
      <Tabs period={period} onChange={setPeriod} />
      <Board period={period} />
      {profile.data && <JoinRanking profile={profile.data} />}
    </BrandScreen>
  );
}
