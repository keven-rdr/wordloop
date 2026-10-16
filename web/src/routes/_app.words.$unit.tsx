import { createFileRoute } from '@tanstack/react-router';
import { DeckPage } from '../features/words/components/DeckPage';

export const Route = createFileRoute('/_app/words/$unit')({ component: DeckRoute });

function DeckRoute() {
  const { unit } = Route.useParams();
  return <DeckPage unitId={unit} />;
}
