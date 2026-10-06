import * as stylex from '@stylexjs/stylex';

export const color = stylex.defineVars({
  bg: { default: '#FFFFFF', '@media (prefers-color-scheme: dark)': '#101820' },
  surface: { default: '#F4F6F8', '@media (prefers-color-scheme: dark)': '#1A2530' },
  text: { default: '#16202A', '@media (prefers-color-scheme: dark)': '#EAF0F5' },
  primary: { default: '#0E9F8E', '@media (prefers-color-scheme: dark)': '#2CC4B2' },
  primaryEdge: { default: '#0A7A6D', '@media (prefers-color-scheme: dark)': '#1B9C8C' },
  success: '#2E9E4F',
  danger: '#D64545',
});

export const space = stylex.defineVars({
  xs: '4px',
  sm: '8px',
  md: '16px',
  lg: '24px',
});

export const radius = stylex.defineVars({
  md: '12px',
  full: '999px',
});
