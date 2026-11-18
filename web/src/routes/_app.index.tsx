import { createFileRoute } from '@tanstack/react-router';
import { HomePage } from '../features/auth/components/HomePage';
import { GoalCard } from '../features/goals/components/GoalCard';
import { ProgressPanel } from '../features/progress/components/ProgressPanel';

// O Início junta os atalhos (auth), a meta (goals) e o painel de progresso (progress): as features não se importam.
export const Route = createFileRoute('/_app/')({
  component: () => (
    <HomePage>
      <GoalCard />
      <ProgressPanel />
    </HomePage>
  ),
});
