import * as stylex from '@stylexjs/stylex';
import { color, radius, size, space } from '../../../shared/ui/tokens/color.stylex';

const focus = {
  outlineStyle: { default: 'none', ':focus-visible': 'solid' },
  outlineWidth: 3,
  outlineColor: color.link,
  outlineOffset: 2,
} as const;

// Montar a frase: a linha de resposta em cima (a frase como a pessoa a montou) e o banco de peças embaixo.
export const tile = stylex.create({
  row: {
    display: 'flex',
    flexWrap: 'wrap',
    alignContent: 'flex-start',
    gap: space.sm,
    minHeight: `calc(${size.touch} * 2 + ${space.sm})`,
    margin: 0,
    padding: space.md,
    listStyle: 'none',
    boxSizing: 'border-box',
    borderRadius: radius.md,
    borderWidth: 2,
    borderStyle: 'dashed',
    borderColor: color.border,
    backgroundColor: color.surface,
  },
  bank: { display: 'flex', flexWrap: 'wrap', gap: space.sm, margin: 0, padding: 0, listStyle: 'none' },
  empty: { margin: 0, color: color.textMuted, fontSize: '1rem' },
  piece: {
    minHeight: size.touch,
    paddingInline: space.md,
    borderRadius: radius.md,
    borderWidth: 2,
    borderStyle: 'solid',
    borderColor: color.border,
    backgroundColor: { default: color.bg, ':hover': color.surface },
    color: color.text,
    fontSize: '1.15rem',
    fontWeight: 600,
    cursor: { default: 'pointer', ':disabled': 'default' },
    ...focus,
  },
  placed: { borderColor: color.primary },
  right: { borderColor: color.success },
  wrong: { borderColor: color.danger },
  key: { marginInlineEnd: space.sm, color: color.textMuted, fontSize: '0.85rem', fontWeight: 700 },
});
