import { createFileRoute } from '@tanstack/react-router';
import { LoginPage } from '../features/auth/components/LoginPage';

const failures = ['failed', 'two_factor', 'email'] as const;
type LoginFailure = (typeof failures)[number];

function parseFailure(value: unknown): LoginFailure | undefined {
  return failures.find((failure) => failure === value);
}

export const Route = createFileRoute('/login')({
  component: LoginPage,
  validateSearch: (search: Record<string, unknown>): { error?: LoginFailure } => ({
    error: parseFailure(search.error),
  }),
});
