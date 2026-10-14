import * as stylex from '@stylexjs/stylex';
import type { ReactNode } from 'react';
import { space } from './tokens/color.stylex';

const styles = stylex.create({
  page: { padding: space.md, maxWidth: '48rem', marginInline: 'auto', width: '100%', boxSizing: 'border-box' },
});

export function AppPage({ children }: Readonly<{ children: ReactNode }>) {
  return <div {...stylex.props(styles.page)}>{children}</div>;
}
