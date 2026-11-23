import * as stylex from '@stylexjs/stylex';
import { brand } from '../tokens/brand.stylex';
import { color, radius, size, space } from '../tokens/color.stylex';
import { bp } from '../tokens/media.stylex';
import { MainNav } from './MainNav';
import type { NavItem } from './types';

const styles = stylex.create({
  bar: {
    position: 'sticky',
    bottom: 0,
    zIndex: 10,
    display: 'flex',
    flexDirection: 'column',
    boxSizing: 'border-box',
    height: `calc(${size.nav} + env(safe-area-inset-bottom))`,
    paddingBottom: 'env(safe-area-inset-bottom)',
    backgroundColor: color.surface,
    borderTopWidth: 1,
    borderTopStyle: 'solid',
    borderTopColor: color.border,
  },
  list: {
    listStyle: 'none',
    margin: 0,
    padding: 0,
    display: 'flex',
    flexGrow: 1,
    width: '100%',
    maxWidth: size.content,
    marginInline: 'auto',
    boxSizing: 'border-box',
  },
  cell: { flexGrow: 1, flexBasis: 0, display: 'flex' },
  item: {
    position: 'relative',
    flexGrow: 1,
    display: 'flex',
    flexDirection: 'column',
    alignItems: 'center',
    justifyContent: 'center',
    gap: space.xs,
    color: color.textMuted,
    fontSize: '0.8rem',
    fontWeight: 700,
    textDecoration: 'none',
    transitionProperty: 'color',
    transitionDuration: { default: '150ms', [bp.reduceMotion]: '0s' },
    outlineStyle: { default: 'none', ':focus-visible': 'solid' },
    outlineWidth: 3,
    outlineColor: brand.yellow,
    outlineOffset: -3,
  },
  active: {
    color: color.link,
    '::before': {
      content: '""',
      position: 'absolute',
      top: 0,
      insetInline: space.lg,
      height: size.marker,
      borderBottomLeftRadius: radius.full,
      borderBottomRightRadius: radius.full,
      backgroundColor: color.link,
    },
  },
  slot: { display: 'flex', alignItems: 'center', justifyContent: 'center', width: '2.25rem', height: '2.25rem' },
  icon: { width: '1.5rem', height: '1.5rem' },
});

type Props = Readonly<{ items: readonly NavItem[]; name: string; pictureUrl?: string }>;

// Menu inferior (mobile e tablet): os itens vêm do registro das features, em ordem.
export function BottomNav(props: Props) {
  const look = {
    bar: styles.bar,
    list: styles.list,
    cell: styles.cell,
    item: styles.item,
    active: styles.active,
    slot: styles.slot,
    icon: styles.icon,
  };
  return <MainNav {...props} look={look} />;
}
