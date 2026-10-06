import * as stylex from '@stylexjs/stylex';
import type { ButtonHTMLAttributes } from 'react';
import { color, radius, space } from './tokens/color.stylex';
import { bp } from './tokens/media.stylex';

const styles = stylex.create({
  button: {
    minHeight: 48,
    paddingInline: space.lg,
    borderRadius: radius.md,
    borderWidth: 0,
    backgroundColor: color.primary,
    color: color.bg,
    fontWeight: 700,
    transitionProperty: 'transform, box-shadow',
    transitionDuration: { default: '120ms', [bp.reduceMotion]: '0s' },
    transform: { default: null, ':active': 'translateY(4px)' },
    boxShadow: { default: `0 4px 0 ${color.primaryEdge}`, ':active': `0 0 0 ${color.primaryEdge}` },
    width: { default: '100%', [bp.desktop]: 320 },
  },
});

export function Button(props: Readonly<ButtonHTMLAttributes<HTMLButtonElement>>) {
  return <button type="button" {...props} {...stylex.props(styles.button)} />;
}
