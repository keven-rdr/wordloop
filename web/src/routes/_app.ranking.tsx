import { createFileRoute } from '@tanstack/react-router';
import { RankingPage } from '../features/leaderboard/components/RankingPage';

export const Route = createFileRoute('/_app/ranking')({ component: RankingPage });
