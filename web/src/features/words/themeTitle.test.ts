import { describe, expect, it } from 'vitest';
import { themeTitle } from './themeTitle';

// `t` de mentira: devolve a chave com os parâmetros que a função decide, ou o texto cru da tradução ausente.
const t = ((key: string, options?: { level?: string; n?: number; defaultValue?: string }) => {
  if (options?.level) return `${key}:${options.level}:${options.n}`;
  return options?.defaultValue ?? key;
}) as never;

describe('themeTitle', () => {
  it('usa a chave única das lições, com nível em maiúsculas e número sem zero à esquerda', () => {
    expect(themeTitle(t, 'lesson-a1-03')).toBe('units.lesson:A1:3');
    expect(themeTitle(t, 'lesson-b1-20')).toBe('units.lesson:B1:20');
  });

  it('mantém os temas antigos com a chave própria e cai para o texto cru se não houver tradução', () => {
    expect(themeTitle(t, 'home-and-family')).toBe('home-and-family');
    expect(themeTitle(t, 'lesson-xx')).toBe('lesson-xx');
  });
});
