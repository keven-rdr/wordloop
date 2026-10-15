import { createFileRoute } from '@tanstack/react-router';
import { ResetPage } from '../features/auth/components/TokenPage';

export const Route = createFileRoute('/reset')({ component: ResetPage });
