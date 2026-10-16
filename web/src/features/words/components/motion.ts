import * as stylex from '@stylexjs/stylex';
import { bp } from '../../../shared/ui/tokens/media.stylex';

export const CARD_MS = 280;

const enter = stylex.keyframes({
  from: { transform: 'translateY(-12px) scale(0.94)', opacity: 0.6 },
  to: { transform: 'none', opacity: 1 },
});

const leaveNext = stylex.keyframes({
  from: { transform: 'none', opacity: 1 },
  to: { transform: 'translateX(-115%) rotate(-14deg)', opacity: 0 },
});

const leavePrevious = stylex.keyframes({
  from: { transform: 'none', opacity: 1 },
  to: { transform: 'translateX(115%) rotate(14deg)', opacity: 0 },
});

export type CardMotion = 'enter' | 'leave-next' | 'leave-previous';

const timing = {
  animationDuration: `${CARD_MS}ms`,
  animationFillMode: 'both',
  animationTimingFunction: 'ease-out',
} as const;

export const motion = stylex.create({
  enter: { ...timing, animationName: { default: enter, [bp.reduceMotion]: 'none' } },
  leaveNext: { ...timing, animationName: { default: leaveNext, [bp.reduceMotion]: 'none' } },
  leavePrevious: { ...timing, animationName: { default: leavePrevious, [bp.reduceMotion]: 'none' } },
  overlay: {
    position: 'absolute',
    inset: 0,
    display: 'flex',
    flexDirection: 'column',
    pointerEvents: 'none',
    zIndex: 2,
  },
});

export const motionStyle = {
  enter: motion.enter,
  'leave-next': motion.leaveNext,
  'leave-previous': motion.leavePrevious,
};

export function reducedMotion(): boolean {
  return typeof window.matchMedia === 'function' && window.matchMedia('(prefers-reduced-motion: reduce)').matches;
}
