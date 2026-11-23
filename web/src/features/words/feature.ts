import { CardsIcon } from '../../shared/ui/nav/Icons';
import type { NavItem } from '../../shared/ui/nav/types';

const nav: readonly NavItem[] = [{ id: 'words', label: 'common:nav.words', icon: CardsIcon, to: '/words', order: 40 }];

export const wordsFeature = { id: 'words', namespaces: ['words'], nav } as const;
