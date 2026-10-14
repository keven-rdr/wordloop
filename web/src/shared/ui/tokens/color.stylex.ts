import * as stylex from '@stylexjs/stylex';

export const color = stylex.defineVars({
  bg: { default: '#FFFFFF', '@media (prefers-color-scheme: dark)': '#120E24' },
  surface: { default: '#F5F2FD', '@media (prefers-color-scheme: dark)': '#1C1636' },
  text: { default: '#1A1530', '@media (prefers-color-scheme: dark)': '#EEEAFB' },
  textMuted: { default: '#5B5478', '@media (prefers-color-scheme: dark)': '#B4ACD3' },
  border: { default: '#DDD6F3', '@media (prefers-color-scheme: dark)': '#3A3060' },
  primary: '#7B5CE0',
  primaryEdge: { default: '#6445C9', '@media (prefers-color-scheme: dark)': '#9277EA' },
  onPrimary: '#FFFFFF',
  link: { default: '#5B3FC4', '@media (prefers-color-scheme: dark)': '#B9A7F7' },
  success: '#2E9E4F',
  danger: { default: '#C23737', '@media (prefers-color-scheme: dark)': '#FF8A8A' },
});

export const space = stylex.defineVars({
  xs: '4px',
  sm: '8px',
  md: '16px',
  lg: '24px',
  xl: '40px',
});

export const radius = stylex.defineVars({
  md: '12px',
  lg: '20px',
  full: '999px',
});

export const size = stylex.defineVars({
  touch: '48px',
  card: '420px',
});
