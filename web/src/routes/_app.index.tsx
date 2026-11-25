import { createFileRoute } from '@tanstack/react-router';
import { HomePage } from '../features/auth/components/HomePage';
import { GoalCard } from '../features/goals/components/GoalCard';
import { ScoreCard } from '../features/leaderboard/components/ScoreCard';
import { ProgressPanel } from '../features/progress/components/ProgressPanel';

// O Início junta os atalhos (auth), a meta (goals), os pontos (leaderboard) e o painel de progresso (progress): as
// features não se importam.
export const Route = createFileRoute('/_app/')({
  component: () => (
    <HomePage goal={<GoalCard />}>
      <ScoreCard />
      <ProgressPanel />
    </HomePage>
  ),
});
