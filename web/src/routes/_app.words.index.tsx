import { createFileRoute } from '@tanstack/react-router';
import { WordsPage } from '../features/words/components/WordsPage';

export const Route = createFileRoute('/_app/words/')({ component: WordsPage });
