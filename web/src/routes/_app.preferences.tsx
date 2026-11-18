import { createFileRoute } from '@tanstack/react-router';
import { PreferencesPage } from '../features/settings/components/PreferencesPage';

export const Route = createFileRoute('/_app/preferences')({ component: PreferencesPage });
