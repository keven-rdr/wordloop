import * as stylex from '@stylexjs/stylex';
import { brand } from '../../../shared/ui/tokens/brand.stylex';
import { space } from '../../../shared/ui/tokens/color.stylex';
import { bp } from '../../../shared/ui/tokens/media.stylex';

const drift = stylex.keyframes({
  '0%': { transform: 'translate(0, 0) scale(1)' },
  '33%': { transform: 'translate(3rem, -2rem) scale(1.12)' },
  '66%': { transform: 'translate(-2rem, 2.5rem) scale(0.94)' },
  '100%': { transform: 'translate(0, 0) scale(1)' },
});

const bob = stylex.keyframes({
  '0%': { transform: 'translateY(0)' },
  '50%': { transform: 'translateY(-0.9rem)' },
  '100%': { transform: 'translateY(0)' },
});

const shimmer = stylex.keyframes({
  '0%': { backgroundPosition: '0% 50%' },
  '50%': { backgroundPosition: '100% 50%' },
  '100%': { backgroundPosition: '0% 50%' },
});

export const art = stylex.create({
  panel: {
    position: 'relative',
    overflow: 'hidden',
    display: 'flex',
    flexDirection: 'column',
    justifyContent: 'space-between',
    gap: space.lg,
    color: brand.onBrand,
    backgroundImage: `linear-gradient(120deg, ${brand.violetNight}, ${brand.violetDeep}, ${brand.violet}, ${brand.violetDeep})`,
    backgroundSize: '300% 300%',
    animationName: { default: shimmer, [bp.reduceMotion]: 'none' },
    animationDuration: '24s',
    animationIterationCount: 'infinite',
    animationTimingFunction: 'ease-in-out',
    minHeight: { default: '7rem', [bp.tablet]: 'auto' },
    padding: { default: space.lg, [bp.tablet]: space.xl },
  },
  orb: {
    position: 'absolute',
    borderRadius: '50%',
    filter: 'blur(3.5rem)',
    opacity: 0.55,
    pointerEvents: 'none',
  },
  orbPink: {
    top: '-4rem',
    right: '-3rem',
    width: '18rem',
    height: '18rem',
    backgroundColor: brand.pink,
    animationName: { default: drift, [bp.reduceMotion]: 'none' },
    animationDuration: '22s',
    animationIterationCount: 'infinite',
    animationTimingFunction: 'ease-in-out',
  },
  orbYellow: {
    bottom: '-5rem',
    left: '-4rem',
    width: '16rem',
    height: '16rem',
    backgroundColor: brand.yellow,
    animationName: { default: drift, [bp.reduceMotion]: 'none' },
    animationDuration: '28s',
    animationIterationCount: 'infinite',
    animationTimingFunction: 'ease-in-out',
  },
  orbViolet: {
    top: '35%',
    left: '40%',
    width: '14rem',
    height: '14rem',
    backgroundColor: brand.violetSoft,
    animationName: { default: drift, [bp.reduceMotion]: 'none' },
    animationDuration: '34s',
    animationIterationCount: 'infinite',
    animationTimingFunction: 'ease-in-out',
  },
  logo: { display: 'flex', alignItems: 'center', gap: space.md, position: 'relative', zIndex: 1 },
  mark: { width: '3rem', height: '3rem', borderRadius: '0.9rem' },
  name: { fontSize: '1.6rem', fontWeight: 700, letterSpacing: '0.01em' },
  stage: {
    position: 'relative',
    zIndex: 1,
    display: { default: 'none', [bp.tablet]: 'flex' },
    justifyContent: 'center',
    alignItems: 'center',
    height: '16rem',
  },
  float: {
    position: 'absolute',
    animationName: { default: bob, [bp.reduceMotion]: 'none' },
    animationDuration: '7s',
    animationIterationCount: 'infinite',
    animationTimingFunction: 'ease-in-out',
  },
  floatB: { animationDelay: '-2.3s', animationDuration: '8.5s' },
  floatC: { animationDelay: '-4.6s', animationDuration: '9.5s' },
  card: {
    width: '9rem',
    height: '12.5rem',
    borderRadius: '1.5rem',
    boxShadow: '0 1.2rem 2.4rem rgba(10, 5, 40, 0.35)',
  },
  cardYellow: { backgroundColor: brand.yellow, transform: 'translateX(-5rem) rotate(-16deg)' },
  cardPink: { backgroundColor: brand.pink, transform: 'translateX(-1rem) rotate(-4deg)' },
  cardPaper: {
    backgroundColor: brand.paper,
    color: brand.violet,
    transform: 'translateX(3rem) rotate(8deg)',
    display: 'flex',
    alignItems: 'center',
    justifyContent: 'center',
  },
  glyph: { width: '5.5rem', height: '5.5rem' },
  copy: { position: 'relative', zIndex: 1, display: { default: 'none', [bp.tablet]: 'block' }, maxWidth: '28rem' },
  headline: { margin: 0, fontSize: '2rem', lineHeight: 1.2, fontWeight: 700 },
  accent: { color: brand.yellow },
  text: { margin: 0, marginTop: space.md, color: brand.onBrandMuted, fontSize: '1.05rem', lineHeight: 1.5 },
});
