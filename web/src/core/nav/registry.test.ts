import { describe, expect, it } from 'vitest';
import { CardsIcon, StudyIcon } from '../../shared/ui/nav/Icons';
import type { NavItem } from '../../shared/ui/nav/types';
import { collectNav, navItems } from './registry';

const item = (id: string, order: number, to: NavItem['to'] = '/study'): NavItem => ({
  id,
  label: `common:nav.${id}`,
  icon: StudyIcon,
  to,
  order,
});

describe('registro do menu', () => {
  it('as features declaram Perfil, Início, Estudar e Palavras, nessa ordem', () => {
    expect(navItems.map((entry) => entry.id)).toEqual(['profile', 'home', 'study', 'words']);
    expect(navItems.map((entry) => entry.to)).toEqual(['/settings', '/', '/study', '/words']);
    expect(navItems[0]?.icon).toBe('avatar');
    expect(navItems[3]?.icon).toBe(CardsIcon);
  });

  it('uma feature nova entra no menu só com o próprio feature.ts, na posição de `order`', () => {
    const modules = {
      '../../features/study/feature.ts': { studyFeature: { id: 'study', nav: [item('study', 30)] } },
      '../../features/leaderboard/feature.ts': {
        leaderboardFeature: { id: 'leaderboard', nav: [{ ...item('ranking', 35), label: 'leaderboard:nav' }] },
      },
      '../../features/words/feature.ts': { wordsFeature: { id: 'words', nav: [item('words', 40, '/words')] } },
    };
    expect(collectNav(modules).map((entry) => entry.id)).toEqual(['study', 'ranking', 'words']);
  });

  it('feature sem menu e exportações que não são feature ficam de fora', () => {
    const modules = {
      a: { goalsFeature: { id: 'goals', namespaces: ['goals'] }, helper: 42, nothing: null },
      b: { studyFeature: { id: 'study', nav: [item('study', 30)] } },
    };
    expect(collectNav(modules).map((entry) => entry.id)).toEqual(['study']);
  });

  it('id repetido é erro de quem declarou', () => {
    const modules = {
      a: { one: { id: 'a', nav: [item('study', 30)] } },
      b: { two: { id: 'b', nav: [item('study', 31)] } },
    };
    expect(() => collectNav(modules)).toThrow('nav: item repetido "study"');
  });
});
