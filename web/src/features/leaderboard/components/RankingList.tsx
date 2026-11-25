import * as stylex from '@stylexjs/stylex';
import { useTranslation } from 'react-i18next';
import type { Leaderboard, LeaderboardEntry } from '../../../api/generated/types.gen';
import { l } from './styles';

function Row({ entry }: Readonly<{ entry: LeaderboardEntry }>) {
  const { t } = useTranslation('leaderboard');
  const name = entry.nickname ?? t('you');
  return (
    <li {...stylex.props(l.row, entry.me && l.rowMe)}>
      <span {...stylex.props(l.place)}>{t('position', { position: entry.position })}</span>
      <span {...stylex.props(l.name)}>{entry.me && entry.nickname ? t('meName', { name }) : name}</span>
      <span {...stylex.props(l.points)}>{t('points', { count: entry.points })}</span>
    </li>
  );
}

// Da 4ª posição em diante; a linha de quem pergunta fica fixa no fim quando não está na lista (ou não participa).
export function RankingList({ board }: Readonly<{ board: Leaderboard }>) {
  const { t } = useTranslation('leaderboard');
  const rest = board.entries.slice(3);
  const meListed = board.entries.some((entry) => entry.me);
  return (
    <>
      {rest.length > 0 && (
        <ol aria-label={t('list')} {...stylex.props(l.list)}>
          {rest.map((entry) => (
            <Row key={`${entry.position}-${entry.nickname}`} entry={entry} />
          ))}
        </ol>
      )}
      {board.me && !meListed && (
        <section aria-label={t('mine')} {...stylex.props(l.section)}>
          <ul {...stylex.props(l.list)}>
            <Row entry={board.me} />
          </ul>
          {!board.me.nickname && <p {...stylex.props(l.text)}>{t('hiddenNote')}</p>}
        </section>
      )}
    </>
  );
}
