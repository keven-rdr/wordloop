import * as stylex from '@stylexjs/stylex';
import type { ReactNode } from 'react';
import { ThemeToggle } from '../../../core/theme/ThemeToggle';
import { BrandPanel } from './BrandPanel';
import { form } from './formStyles';

export function AuthShell({ children }: Readonly<{ children: ReactNode }>) {
  return (
    <div {...stylex.props(form.shell)}>
      <BrandPanel />
      <section {...stylex.props(form.side)}>
        <div {...stylex.props(form.bar)}>
          <ThemeToggle />
        </div>
        <div {...stylex.props(form.center)}>
          <div {...stylex.props(form.card)}>{children}</div>
        </div>
      </section>
    </div>
  );
}
