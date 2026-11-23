import { describe, expect, it } from 'vitest';
import { desktopQuery } from './breakpoints';
import source from './media.stylex.ts?raw';

describe('breakpoints', () => {
  it('a consulta do JS é a mesma largura de media.stylex.ts', () => {
    expect(source).toContain(`desktop: '@media ${desktopQuery}'`);
  });
});
