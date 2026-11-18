import { createFileRoute } from '@tanstack/react-router';
import { SessionPage } from '../features/study/components/SessionPage';

export const Route = createFileRoute('/_app/study/session/$id')({ component: SessionRoute });

function SessionRoute() {
  const { id } = Route.useParams();
  return <SessionPage sessionId={id} />;
}
