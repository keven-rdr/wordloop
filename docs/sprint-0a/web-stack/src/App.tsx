import * as stylex from '@stylexjs/stylex';
import { useState } from 'react';
import { Button } from './ui/Button';
import { ProgressBar } from './ui/ProgressBar';
import { ThemeSwitch } from './ui/ThemeSwitch';
import { color, space } from './ui/tokens/color.stylex';
import { darkTheme, lightTheme } from './ui/tokens/themes.stylex';

const styles = stylex.create({
  page: { minHeight: '100vh', backgroundColor: color.bg, color: color.text, padding: space.md },
});

export function App() {
  const [dark, setDark] = useState(false);
  const [pct, setPct] = useState(20);
  return (
    <div {...stylex.props(styles.page, dark ? darkTheme : lightTheme)} data-testid="page">
      <h1>wordloop spike</h1>
      <ThemeSwitch checked={dark} onChange={setDark} label="Tema escuro" />
      <ProgressBar value={pct} />
      <Button onClick={() => setPct((p) => Math.min(100, p + 20))}>Continuar</Button>
    </div>
  );
}
