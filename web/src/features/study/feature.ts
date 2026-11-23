import { StudyIcon } from '../../shared/ui/nav/Icons';
import type { NavItem } from '../../shared/ui/nav/types';

const nav: readonly NavItem[] = [{ id: 'study', label: 'common:nav.study', icon: StudyIcon, to: '/study', order: 30 }];

export const studyFeature = { id: 'study', namespaces: ['study'], nav } as const;
