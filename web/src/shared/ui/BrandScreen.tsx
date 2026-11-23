import * as stylex from '@stylexjs/stylex';
import type { ReactNode } from 'react';
import { brand } from './tokens/brand.stylex';
import { color, radius, size, space } from './tokens/color.stylex';
import { bp } from './tokens/media.stylex';

const styles = stylex.create({
  screen: {
    flexGrow: 1,
    display: 'flex',
    flexDirection: 'column',
    backgroundColor: brand.violetDeep,
    color: brand.onBrand,
  },
  inner: { width: '100%', maxWidth: size.content, marginInline: 'auto', boxSizing: 'border-box' },
  wide: { maxWidth: { default: size.content, [bp.desktop]: size.wide } },
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

type Props = Readonly<{ title: string; subtitle?: string; lead?: ReactNode; wide?: boolean; children: ReactNode }>;

// Cabeçalho roxo da marca e folha de conteúdo arredondada, como na tela Palavras. `wide`: no desktop o conteúdo pode
// ir até `size.wide` (para grades como a do Início); no mobile e no tablet nada muda.
export function BrandScreen({ title, subtitle, lead, wide, children }: Props) {
  return (
    <div {...stylex.props(styles.screen)}>
      <header {...stylex.props(styles.inner, wide && styles.wide, styles.header)}>
        {lead}
        <div {...stylex.props(styles.texts)}>
          <h1 {...stylex.props(styles.title)}>{title}</h1>
          {subtitle && <p {...stylex.props(styles.subtitle)}>{subtitle}</p>}
        </div>
      </header>
      <div {...stylex.props(styles.sheet)}>
        <div {...stylex.props(styles.inner, wide && styles.wide, styles.body)}>{children}</div>
      </div>
    </div>
  );
}
