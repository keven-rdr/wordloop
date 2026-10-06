import * as stylex from '@stylexjs/stylex';
import { color } from './color.stylex';

export const lightTheme = stylex.createTheme(color, {
  bg: '#FFFFFF',
  surface: '#F4F6F8',
  text: '#16202A',
  primary: '#0E9F8E',
  primaryEdge: '#0A7A6D',
});

export const darkTheme = stylex.createTheme(color, {
  bg: '#101820',
  surface: '#1A2530',
  text: '#EAF0F5',
  primary: '#2CC4B2',
  primaryEdge: '#1B9C8C',
});
