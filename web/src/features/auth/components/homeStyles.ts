import * as stylex from '@stylexjs/stylex';
import { space } from '../../../shared/ui/tokens/color.stylex';
import { bp } from '../../../shared/ui/tokens/media.stylex';

// Início: uma coluna no mobile (atalhos, meta, painel); no desktop, atalhos e meta à esquerda e o painel ao lado.
export const home = stylex.create({
  grid: {
    display: 'grid',
    gridTemplateColumns: { default: 'minmax(0, 1fr)', [bp.desktop]: 'minmax(0, 1fr) minmax(0, 2fr)' },
    alignItems: 'start',
    gap: { default: space.md, [bp.desktop]: space.lg },
  },
  stack: { display: 'flex', flexDirection: 'column', gap: space.md, minWidth: 0 },
  main: { minWidth: 0 },
});
