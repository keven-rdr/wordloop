import { HomeIcon } from '../../shared/ui/nav/Icons';
import type { NavItem } from '../../shared/ui/nav/types';

// Perfil (configurações e passkeys) e Início são telas desta feature.
const nav: readonly NavItem[] = [
  { id: 'profile', label: 'common:nav.profile', icon: 'avatar', to: '/settings', order: 10, also: ['/passkeys'] },
  { id: 'home', label: 'common:nav.home', icon: HomeIcon, to: '/', order: 20 },
];

export const authFeature = { id: 'auth', namespaces: ['auth'], nav } as const;
