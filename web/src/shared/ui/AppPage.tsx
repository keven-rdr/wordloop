import * as stylex from '@stylexjs/stylex';
import type { ReactNode } from 'react';
import { size, space } from './tokens/color.stylex';
import { bp } from './tokens/media.stylex';

const styles = stylex.create({
  page: { padding: space.md, maxWidth: size.page, marginInline: 'auto', width: '100%', boxSizing: 'border-box' },
  wide: { maxWidth: { default: size.page, [bp.desktop]: size.wide } },
});

// `wide`: no desktop a página pode ir até `size.wide` (no mobile e no tablet nada muda).
export function AppPage({ children, wide }: Readonly<{ children: ReactNode; wide?: boolean }>) {
  return <div {...stylex.props(styles.page, wide && styles.wide)}>{children}</div>;
}
