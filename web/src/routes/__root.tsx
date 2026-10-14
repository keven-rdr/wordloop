import * as stylex from '@stylexjs/stylex';
import { createRootRoute, Outlet } from '@tanstack/react-router';
import { NotFound } from '../core/NotFound';
import { VersionFooter } from '../core/VersionFooter';
import { color, space } from '../shared/ui/tokens/color.stylex';

const styles = stylex.create({
  page: {
    minHeight: '100vh',
    display: 'flex',
    flexDirection: 'column',
    backgroundColor: color.bg,
    color: color.text,
  },
  main: { flexGrow: 1, padding: space.md },
});

export const Route = createRootRoute({ component: RootLayout, notFoundComponent: NotFound });

function RootLayout() {
  return (
    <div {...stylex.props(styles.page)}>
      <main {...stylex.props(styles.main)}>
        <Outlet />
      </main>
      <VersionFooter />
    </div>
  );
}
