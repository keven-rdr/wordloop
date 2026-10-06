import * as stylex from '@stylexjs/stylex';
import { useTranslation } from 'react-i18next';
import { VersionFooter } from './core/VersionFooter';
import { color, space } from './shared/ui/tokens/color.stylex';

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

export function App() {
  const { t } = useTranslation('common');
  return (
    <div {...stylex.props(styles.page)}>
      <main {...stylex.props(styles.main)}>
        <h1>{t('app.name')}</h1>
        <p>{t('app.tagline')}</p>
      </main>
      <VersionFooter />
    </div>
  );
}
