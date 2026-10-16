import { createFileRoute } from '@tanstack/react-router';
import { WordsPage } from '../features/words/components/WordsPage';

export const Route = createFileRoute('/words/')({ component: WordsPage });
