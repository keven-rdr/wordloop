import { createFileRoute } from '@tanstack/react-router';
import { PlanPage } from '../features/study/components/PlanPage';

export const Route = createFileRoute('/_app/study/')({ component: PlanPage });
