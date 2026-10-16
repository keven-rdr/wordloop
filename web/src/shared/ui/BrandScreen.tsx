import * as stylex from '@stylexjs/stylex';
import type { ReactNode } from 'react';
import { brand } from './tokens/brand.stylex';
import { color, radius, space } from './tokens/color.stylex';

const styles = stylex.create({
  screen: {
    flexGrow: 1,
    display: 'flex',
    flexDirection: 'column',
    backgroundColor: brand.violetDeep,
    color: brand.onBrand,
  },
  inner: { width: '100%', maxWidth: '40rem', marginInline: 'auto', boxSizing: 'border-box' },
  header: { display: 'flex', alignItems: 'center', gap: space.md, padding: space.lg },
  texts: { display: 'flex', flexDirection: 'column', gap: space.xs, minWidth: 0 },
  title: { margin: 0, fontSize: '1.9rem', lineHeight: 1.1, fontWeight: 800, overflowWrap: 'anywhere' },
  subtitle: { margin: 0, color: brand.onBrandMuted, overflowWrap: 'anywhere' },
  sheet: {
    flexGrow: 1,
    display: 'flex',
    flexDirection: 'column',
    backgroundColor: color.bg,
    color: color.text,
    borderTopLeftRadius: radius.lg,
    borderTopRightRadius: radius.lg,
  },
  body: { display: 'flex', flexDirection: 'column', gap: space.md, padding: space.lg },
});

type Props = Readonly<{ title: string; subtitle?: string; lead?: ReactNode; children: ReactNode }>;

// Cabeçalho roxo da marca e folha de conteúdo arredondada, como na tela Palavras.
export function BrandScreen({ title, subtitle, lead, children }: Props) {
  return (
    <div {...stylex.props(styles.screen)}>
      <header {...stylex.props(styles.inner, styles.header)}>
        {lead}
        <div {...stylex.props(styles.texts)}>
          <h1 {...stylex.props(styles.title)}>{title}</h1>
          {subtitle && <p {...stylex.props(styles.subtitle)}>{subtitle}</p>}
        </div>
      </header>
      <div {...stylex.props(styles.sheet)}>
        <div {...stylex.props(styles.inner, styles.body)}>{children}</div>
      </div>
    </div>
  );
}
