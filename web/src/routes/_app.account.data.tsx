import { createFileRoute } from '@tanstack/react-router';
import { DataPage } from '../features/privacy/components/DataPage';

export const Route = createFileRoute('/_app/account/data')({ component: DataPage });
