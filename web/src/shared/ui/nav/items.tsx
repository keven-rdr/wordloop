import { useRouterState } from '@tanstack/react-router';
import type { CSSProperties } from 'react';
import { useTranslation } from 'react-i18next';
import { Avatar } from './Avatar';
import type { NavItem } from './types';

const inside = (path: string, base: string) => path === base || path.startsWith(`${base}/`);

// O item está marcado na rota dele, nas rotas abaixo dela e nas rotas de `also` (o Início só em `/`).
export function isActive(item: NavItem, path: string): boolean {
  const bases = [String(item.to), ...(item.also ?? [])];
  return bases.some((base) => (base === '/' ? path === '/' : inside(path, base)));
}

// Namespace de cada rótulo (`common:nav.home` → `common`); sem prefixo, `common`.
export function labelNamespaces(items: readonly NavItem[]): string[] {
  const all = items.map((item) => /^([^:]+):/.exec(item.label)?.[1] ?? 'common');
  return [...new Set(['common', ...all])];
}

// Rótulos traduzidos e o item ativo, para os dois menus (inferior e lateral).
export function useNavState(items: readonly NavItem[]) {
  const { t } = useTranslation(labelNamespaces(items));
  const path = useRouterState({ select: (state) => state.location.pathname });
  return {
    title: t('common:nav.label'),
    appName: t('common:app.name'),
    label: (item: NavItem) => t(item.label),
    active: (item: NavItem) => isActive(item, path),
  };
}

type GlyphProps = Readonly<{
  item: NavItem;
  name: string;
  pictureUrl?: string;
  className?: string;
  style?: CSSProperties;
}>;

// O ícone do item, ou a foto (inicial) do usuário no Perfil.
export function NavGlyph({ item, name, pictureUrl, className, style }: GlyphProps) {
  if (item.icon === 'avatar') return <Avatar name={name} pictureUrl={pictureUrl} />;
  const Icon = item.icon;
  return <Icon className={className} style={style} />;
}
