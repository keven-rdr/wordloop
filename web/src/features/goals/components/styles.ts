import * as stylex from '@stylexjs/stylex';
import { color, radius, size, space } from '../../../shared/ui/tokens/color.stylex';

// O comum (cartao, textos, botao, campo) vem de shared/ui/panelStyles; aqui fica o que e so das metas.
export const g = stylex.create({
  track: { height: 12, borderRadius: radius.full, backgroundColor: color.border, overflow: 'hidden' },
  fill: (percent: number) => ({
    width: `${percent}%`,
    height: '100%',
    borderRadius: radius.full,
    backgroundColor: color.primary,
  }),
  stats: { display: 'flex', gap: space.sm, margin: 0, padding: 0, listStyle: 'none' },
  stat: {
    flexBasis: 0,
    flexGrow: 1,
    display: 'flex',
    flexDirection: 'column',
    alignItems: 'center',
    gap: space.xs,
    padding: space.md,
    borderRadius: radius.md,
    backgroundColor: color.surface,
  },
  statValue: { fontSize: '1.6rem', fontWeight: 800 },
  statLabel: { color: color.textMuted, fontSize: '0.85rem', textAlign: 'center' },
});

// Escolha da meta: cada opção é um cartão com um radio de verdade dentro (teclado e leitor de tela de graça).
export const c = stylex.create({
  options: { display: 'flex', flexDirection: 'column', gap: space.sm, margin: 0, padding: 0, borderWidth: 0 },
  legend: { padding: 0, marginBottom: space.sm, fontWeight: 700 },
  option: {
    display: 'flex',
    alignItems: 'center',
    gap: space.md,
    minHeight: size.touch,
    padding: space.md,
    borderRadius: radius.md,
    borderWidth: 2,
    borderStyle: 'solid',
    borderColor: color.border,
    backgroundColor: color.bg,
    cursor: 'pointer',
  },
  chosen: { borderColor: color.primary, backgroundColor: color.surface },
  radio: { width: '1.25rem', height: '1.25rem', flexShrink: 0, accentColor: color.primary },
  optionText: { display: 'flex', flexDirection: 'column', gap: space.xs },
  optionName: { fontWeight: 700 },
  fields: { display: 'flex', gap: space.sm },
  field: { flexBasis: 0, flexGrow: 1, display: 'flex', flexDirection: 'column', gap: space.xs },
  block: { width: '100%', boxSizing: 'border-box' },
});
