import * as stylex from '@stylexjs/stylex';
import { color, radius, size, space } from '../../../shared/ui/tokens/color.stylex';

// Foco visível do teclado nas abas (o StyleX só aceita valores deste arquivo dentro do stylex.create).
const focusRing = {
  outlineStyle: { default: 'none', ':focus-visible': 'solid' },
  outlineWidth: 3,
  outlineColor: color.link,
  outlineOffset: 2,
} as const;

// O que é só do ranking (o comum vem de shared/ui/panelStyles): abas, lista, participação e pódio.
export const l = stylex.create({
  tabs: { display: 'flex', gap: space.sm, margin: 0, padding: 0, borderWidth: 0 },
  tab: {
    flexBasis: 0,
    flexGrow: 1,
    minHeight: size.touch,
    borderRadius: radius.md,
    borderWidth: 2,
    borderStyle: 'solid',
    borderColor: color.border,
    backgroundColor: color.bg,
    color: color.text,
    fontSize: '1rem',
    fontWeight: 700,
    cursor: 'pointer',
    ...focusRing,
  },
  tabOn: { borderColor: color.primary, backgroundColor: color.surface },
  list: { display: 'flex', flexDirection: 'column', gap: space.xs, margin: 0, padding: 0, listStyle: 'none' },
  row: {
    display: 'flex',
    alignItems: 'center',
    gap: space.md,
    minHeight: size.touch,
    paddingInline: space.md,
    borderRadius: radius.md,
    backgroundColor: color.surface,
  },
  rowMe: { borderWidth: 2, borderStyle: 'solid', borderColor: color.primary },
  place: { minWidth: space.xl, fontWeight: 800, color: color.textMuted },
  name: { flexGrow: 1, fontWeight: 700, overflowWrap: 'anywhere' },
  points: { fontWeight: 800 },
  check: { display: 'flex', alignItems: 'center', gap: space.sm, minHeight: size.touch, fontWeight: 600 },
  box: { width: '1.25rem', height: '1.25rem', accentColor: color.primary },
});

// Pódio: três degraus (2º, 1º, 3º), o do meio mais alto.
export const p = stylex.create({
  podium: {
    display: 'flex',
    alignItems: 'flex-end',
    justifyContent: 'center',
    gap: space.sm,
    margin: 0,
    padding: 0,
    listStyle: 'none',
  },
  step: {
    flexBasis: 0,
    flexGrow: 1,
    display: 'flex',
    flexDirection: 'column',
    alignItems: 'center',
    justifyContent: 'flex-end',
    gap: space.xs,
    paddingBlock: space.md,
    paddingInline: space.xs,
    borderTopLeftRadius: radius.md,
    borderTopRightRadius: radius.md,
    color: color.onMedal,
    textAlign: 'center',
  },
  first: { minHeight: '9rem', backgroundColor: color.gold },
  second: { minHeight: '7rem', backgroundColor: color.silver },
  third: { minHeight: '5.5rem', backgroundColor: color.bronze },
  medal: { fontSize: '1.6rem', fontWeight: 800 },
  who: { fontWeight: 800, overflowWrap: 'anywhere' },
  pts: { fontSize: '0.9rem', fontWeight: 600 },
});
