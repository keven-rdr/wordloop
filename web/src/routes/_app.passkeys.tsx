import { createFileRoute } from '@tanstack/react-router';
import { PasskeysPage } from '../features/auth/components/PasskeysPage';

export const Route = createFileRoute('/_app/passkeys')({ component: PasskeysPage });
