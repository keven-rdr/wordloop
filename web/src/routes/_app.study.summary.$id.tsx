import { createFileRoute } from '@tanstack/react-router';
import { SummaryPage } from '../features/study/components/SummaryPage';

export const Route = createFileRoute('/_app/study/summary/$id')({ component: SummaryRoute });

function SummaryRoute() {
  const { id } = Route.useParams();
  return <SummaryPage sessionId={id} />;
}
