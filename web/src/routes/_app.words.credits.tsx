import { createFileRoute } from '@tanstack/react-router';
import { CreditsPage } from '../features/words/components/CreditsPage';

export const Route = createFileRoute('/_app/words/credits')({ component: CreditsPage });
