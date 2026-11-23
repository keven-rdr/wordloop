import type { LinkProps } from '@tanstack/react-router';
import type { ComponentType, SVGProps } from 'react';

export type NavIconComponent = ComponentType<Readonly<SVGProps<SVGSVGElement>>>;

// Um item do menu principal, declarado pela própria feature no `feature.ts` (o menu não tem lista central).
export type NavItem = Readonly<{
  id: string;
  // Chave i18n com namespace, ex.: `common:nav.home` ou `leaderboard:nav`.
  label: string;
  // Ícone SVG, ou `avatar` para a foto do usuário (Perfil).
  icon: NavIconComponent | 'avatar';
  to: NonNullable<LinkProps['to']>;
  // Posição no menu (menor primeiro); deixe folga entre as features (10, 20, 30...).
  order: number;
  // Outras rotas em que o item fica marcado (além de `to` e das rotas abaixo dela).
  also?: readonly string[];
}>;

// O que o registro lê de cada `features/*/feature.ts`.
export type FeatureNav = Readonly<{ id: string; nav?: readonly NavItem[] }>;
