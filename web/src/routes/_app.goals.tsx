import { createFileRoute } from '@tanstack/react-router';
import { GoalsPage } from '../features/goals/components/GoalsPage';

// from=study: a meta foi pedida antes de estudar; salva a meta, volta para Estudar.
type GoalsSearch = { from?: 'study' };

export const Route = createFileRoute('/_app/goals')({
  component: GoalsPage,
  validateSearch: (search: Record<string, unknown>): GoalsSearch => (search.from === 'study' ? { from: 'study' } : {}),
});
