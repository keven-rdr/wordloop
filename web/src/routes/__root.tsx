import * as stylex from '@stylexjs/stylex';
import { createRootRoute, Outlet } from '@tanstack/react-router';
import { NotFound } from '../core/NotFound';
import { useThemeChoice } from '../core/theme/useTheme';
import { color } from '../shared/ui/tokens/color.stylex';
import { darkTheme, lightTheme } from '../shared/ui/tokens/themes.stylex';

const styles = stylex.create({
  page: {
    minHeight: '100vh',
    display: 'flex',
    flexDirection: 'column',
    backgroundColor: color.bg,
    color: color.text,
  },
  main: { flexGrow: 1, display: 'flex', flexDirection: 'column' },
});

export const Route = createRootRoute({ component: RootLayout, notFoundComponent: NotFound });

const themes = { system: undefined, light: lightTheme, dark: darkTheme };

function RootLayout() {
  const theme = themes[useThemeChoice()];
  return (
    <div {...stylex.props(styles.page, theme)}>
      <main {...stylex.props(styles.main)}>
        <Outlet />
      </main>
    </div>
  );
}
