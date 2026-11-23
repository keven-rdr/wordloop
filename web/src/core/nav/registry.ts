import type { FeatureNav, NavItem } from '../../shared/ui/nav/types';

type FeatureModule = Readonly<Record<string, unknown>>;

const isFeature = (value: unknown): value is FeatureNav =>
  typeof value === 'object' && value !== null && 'id' in value && typeof value.id === 'string';

// Junta os itens de menu de todos os módulos `feature.ts`, em ordem. Id repetido é erro de quem declarou.
export function collectNav(modules: Readonly<Record<string, FeatureModule>>): NavItem[] {
  const items = Object.values(modules)
    .flatMap((mod) => Object.values(mod).filter(isFeature))
    .flatMap((feature) => feature.nav ?? []);
  const ids = new Set<string>();
  for (const item of items) {
    if (ids.has(item.id)) throw new Error(`nav: item repetido "${item.id}"`);
    ids.add(item.id);
  }
  return [...items].sort((a, b) => a.order - b.order);
}

// Cada feature declara o próprio menu em `features/<nome>/feature.ts`; uma feature nova entra no menu sem editar
// arquivo compartilhado.
const modules = import.meta.glob<FeatureModule>('../../features/*/feature.ts', { eager: true });

export const navItems: readonly NavItem[] = collectNav(modules);
