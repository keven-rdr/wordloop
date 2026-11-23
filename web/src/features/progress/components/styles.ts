import * as stylex from '@stylexjs/stylex';
import { color, radius, space } from '../../../shared/ui/tokens/color.stylex';
import { bp } from '../../../shared/ui/tokens/media.stylex';

// Cartões do painel: simples, de baixo contraste de borda, em coluna única (mobile-first); no desktop, duas colunas.
export const p = stylex.create({
  panel: {
    display: { default: 'flex', [bp.desktop]: 'grid' },
    flexDirection: 'column',
    gridTemplateColumns: { default: 'none', [bp.desktop]: 'repeat(2, minmax(0, 1fr))' },
    alignItems: { default: 'stretch', [bp.desktop]: 'start' },
    gap: space.md,
  },
  // Título e convite ocupam a linha toda da grade (no mobile não muda nada).
  full: { gridColumn: '1 / -1' },
  card: {
    display: 'flex',
    flexDirection: 'column',
    gap: space.sm,
    padding: space.md,
    borderRadius: radius.md,
    borderWidth: 1,
    borderStyle: 'solid',
    borderColor: color.border,
    backgroundColor: color.surface,
  },
  title: { margin: 0, fontSize: '0.95rem', fontWeight: 700, color: color.textMuted },
  big: { margin: 0, fontSize: '1.8rem', fontWeight: 800, lineHeight: 1.1 },
  text: { margin: 0, color: color.textMuted, fontSize: '0.95rem' },
  row: { display: 'flex', gap: space.sm },
  half: { flexBasis: 0, flexGrow: 1, minWidth: 0 },
  muted: { color: color.textMuted },
  heading: { margin: 0, fontSize: '1.2rem' },
  error: { margin: 0, color: color.danger },
});

// Heatmap: 7 linhas (segunda a domingo) e uma coluna por semana; o tom é a opacidade do roxo.
export const h = stylex.create({
  grid: {
    display: 'grid',
    gridTemplateRows: 'repeat(7, 1fr)',
    gridAutoFlow: 'column',
    gridAutoColumns: '1fr',
    gap: 3,
  },
  cell: { aspectRatio: '1 / 1', borderRadius: 3, backgroundColor: color.border },
  hidden: { visibility: 'hidden' },
  level: (opacity: number) => ({ backgroundColor: color.primary, opacity }),
});

// Previsão: 7 colunas, a altura da barra vem de uma função (valor dinâmico).
export const f = stylex.create({
  bars: { display: 'flex', alignItems: 'flex-end', gap: space.xs, height: '5rem' },
  col: {
    flexBasis: 0,
    flexGrow: 1,
    display: 'flex',
    flexDirection: 'column',
    justifyContent: 'flex-end',
    height: '100%',
  },
  bar: (percent: number) => ({
    height: `${percent}%`,
    minHeight: percent > 0 ? 4 : 0,
    borderRadius: 3,
    backgroundColor: color.primary,
  }),
  labels: { display: 'flex', gap: space.xs },
  label: { flexBasis: 0, flexGrow: 1, textAlign: 'center', fontSize: '0.75rem', color: color.textMuted },
});

export const m = stylex.create({
  list: { listStyle: 'none', margin: 0, padding: 0, display: 'flex', flexDirection: 'column', gap: space.sm },
  item: { display: 'flex', alignItems: 'baseline', justifyContent: 'space-between', gap: space.md },
  term: { fontWeight: 700 },
  gloss: { color: color.textMuted },
  misses: { color: color.danger, fontSize: '0.9rem', fontWeight: 600, whiteSpace: 'nowrap' },
});
