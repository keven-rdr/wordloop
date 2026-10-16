import { createFileRoute } from '@tanstack/react-router';
import { HomePage } from '../features/auth/components/HomePage';

export const Route = createFileRoute('/')({ component: HomePage });
