import { createFileRoute } from '@tanstack/react-router';
import { VerifyPage } from '../features/auth/components/TokenPage';

export const Route = createFileRoute('/verify')({ component: VerifyPage });
