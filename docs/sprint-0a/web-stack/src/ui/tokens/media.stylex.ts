import * as stylex from '@stylexjs/stylex';

export const bp = stylex.defineConsts({
  tablet: '@media (min-width: 768px)',
  desktop: '@media (min-width: 1100px)',
  reduceMotion: '@media (prefers-reduced-motion: reduce)',
});
