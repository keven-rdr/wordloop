import * as stylex from '@stylexjs/stylex';
import { color, radius, space } from '../../../shared/ui/tokens/color.stylex';

export const ui = stylex.create({
  panel: { display: 'flex', flexDirection: 'column', alignItems: 'flex-start', gap: space.md, marginTop: space.lg },
  button: {
    display: 'inline-flex',
    alignItems: 'center',
    justifyContent: 'center',
    minHeight: '48px',
    minWidth: '48px',
    paddingInline: space.lg,
    borderRadius: radius.md,
    borderWidth: 0,
    backgroundColor: { default: color.primary, ':hover': color.primaryEdge },
    color: color.onPrimary,
    fontSize: '1rem',
    fontWeight: 600,
    textDecoration: 'none',
    cursor: 'pointer',
  },
  secondary: {
    backgroundColor: { default: color.surface, ':hover': color.surface },
    color: color.text,
    borderWidth: 1,
    borderStyle: 'solid',
    borderColor: color.primaryEdge,
  },
  error: { color: color.danger, margin: 0 },
  muted: { opacity: 0.75, margin: 0 },
  list: {
    listStyle: 'none',
    padding: 0,
    margin: 0,
    display: 'flex',
    flexDirection: 'column',
    gap: space.md,
    width: '100%',
  },
  item: {
    display: 'flex',
    flexDirection: 'column',
    gap: space.sm,
    padding: space.md,
    borderRadius: radius.md,
    backgroundColor: color.surface,
  },
});
