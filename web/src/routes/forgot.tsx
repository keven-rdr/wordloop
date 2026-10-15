import { createFileRoute } from '@tanstack/react-router';
import { ForgotPage } from '../features/auth/components/ForgotPage';

export const Route = createFileRoute('/forgot')({ component: ForgotPage });
