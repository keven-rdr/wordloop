import * as stylex from '@stylexjs/stylex';
import { color, radius, size, space } from '../../../shared/ui/tokens/color.stylex';

const focus = {
  outlineStyle: { default: 'none', ':focus-visible': 'solid' },
  outlineWidth: 3,
  outlineColor: color.link,
  outlineOffset: 2,
} as const;

export const p = stylex.create({
  section: {
    display: 'flex',
    flexDirection: 'column',
    gap: space.sm,
    margin: 0,
    padding: 0,
    borderWidth: 0,
  },
  legend: { padding: 0, marginBottom: space.sm, fontSize: '1.1rem', fontWeight: 800 },
  hint: { margin: 0, color: color.textMuted, fontSize: '0.95rem' },
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
  control: { width: '1.25rem', height: '1.25rem', flexShrink: 0, accentColor: color.primary },
  name: { fontWeight: 700 },
  texts: { display: 'flex', flexDirection: 'column', gap: space.xs },
  select: {
    minHeight: size.touch,
    paddingInline: space.md,
    borderRadius: radius.md,
    borderWidth: 1,
    borderStyle: 'solid',
    borderColor: color.border,
    backgroundColor: color.bg,
    color: color.text,
    fontSize: '1rem',
    width: '100%',
    boxSizing: 'border-box',
    ...focus,
  },
  label: { fontSize: '0.9rem', fontWeight: 600 },
  field: { display: 'flex', flexDirection: 'column', gap: space.xs },
  status: { margin: 0, minHeight: '1.5rem', color: color.textMuted },
  error: { margin: 0, color: color.danger, fontWeight: 600 },
  stack: { display: 'flex', flexDirection: 'column', gap: space.lg },
});
