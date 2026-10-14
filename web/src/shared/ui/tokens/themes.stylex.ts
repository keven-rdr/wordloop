import * as stylex from '@stylexjs/stylex';
import { color } from './color.stylex';

export const lightTheme = stylex.createTheme(color, {
  bg: '#FFFFFF',
  surface: '#F5F2FD',
  text: '#1A1530',
  textMuted: '#5B5478',
  border: '#DDD6F3',
  primaryEdge: '#6445C9',
  link: '#5B3FC4',
  danger: '#C23737',
});

export const darkTheme = stylex.createTheme(color, {
  bg: '#120E24',
  surface: '#1C1636',
  text: '#EEEAFB',
  textMuted: '#B4ACD3',
  border: '#3A3060',
  primaryEdge: '#9277EA',
  link: '#B9A7F7',
  danger: '#FF8A8A',
});
