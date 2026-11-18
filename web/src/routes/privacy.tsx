import { createFileRoute } from '@tanstack/react-router';
import { LegalPage } from '../features/privacy/components/LegalPage';

export const Route = createFileRoute('/privacy')({ component: () => <LegalPage kind="privacy" /> });
