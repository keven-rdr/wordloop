import { createFileRoute } from '@tanstack/react-router';
import { AppLayout } from '../core/AppLayout';

export const Route = createFileRoute('/_app')({ component: AppLayout });
