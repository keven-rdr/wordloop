import type { NavItem } from '../../shared/ui/nav/types';
import { TrophyIcon } from './components/TrophyIcon';

// Ranking entre Estudar (30) e Palavras (40).
const nav: readonly NavItem[] = [
  { id: 'ranking', label: 'common:nav.ranking', icon: TrophyIcon, to: '/ranking', order: 35 },
];

export const leaderboardFeature = { id: 'leaderboard', namespaces: ['leaderboard'], nav } as const;
