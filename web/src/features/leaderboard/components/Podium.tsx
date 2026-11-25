import * as stylex from '@stylexjs/stylex';
import { useTranslation } from 'react-i18next';
import type { LeaderboardEntry } from '../../../api/generated/types.gen';
import { p } from './styles';

const steps = { 1: p.first, 2: p.second, 3: p.third } as const;

// Os três primeiros em degraus: 2º à esquerda, 1º no meio (mais alto), 3º à direita. Com empate, o lugar se repete.
export function Podium({ entries }: Readonly<{ entries: readonly LeaderboardEntry[] }>) {
  const { t } = useTranslation('leaderboard');
  const top = entries.slice(0, 3);
  const order = [top[1], top[0], top[2]].filter((e): e is LeaderboardEntry => e !== undefined);
  return (
    <ol aria-label={t('podium.label')} {...stylex.props(p.podium)}>
      {order.map((entry) => {
        const place = Math.min(3, entry.position) as 1 | 2 | 3;
        return (
          <li key={`${entry.position}-${entry.nickname}`} {...stylex.props(p.step, steps[place])}>
            <span {...stylex.props(p.medal)}>{t('position', { position: entry.position })}</span>
            <span {...stylex.props(p.who)}>{entry.me ? t('meName', { name: entry.nickname }) : entry.nickname}</span>
            <span {...stylex.props(p.pts)}>{t('points', { count: entry.points })}</span>
          </li>
        );
      })}
    </ol>
  );
}
