import * as stylex from '@stylexjs/stylex';
import { Link } from '@tanstack/react-router';
import { brand } from '../tokens/brand.stylex';
import { color, radius, size, space } from '../tokens/color.stylex';
import { bp } from '../tokens/media.stylex';
import { NavGlyph, useNavState } from './items';
import type { NavItem } from './types';

const styles = stylex.create({
  bar: {
    position: 'fixed',
    insetBlock: 0,
    insetInlineStart: 0,
    zIndex: 10,
    display: 'flex',
    flexDirection: 'column',
    gap: space.lg,
    boxSizing: 'border-box',
    width: size.sideNav,
    padding: space.md,
    overflowY: 'auto',
    backgroundColor: color.surface,
    borderInlineEndWidth: 1,
    borderInlineEndStyle: 'solid',
    borderInlineEndColor: color.border,
  },
  brand: {
    margin: 0,
    paddingInline: space.sm,
    paddingBlock: space.sm,
    color: color.link,
    fontSize: '1.6rem',
    fontWeight: 800,
    lineHeight: 1,
  },
  list: { listStyle: 'none', margin: 0, padding: 0, display: 'flex', flexDirection: 'column', gap: space.xs },
  item: {
    display: 'flex',
    alignItems: 'center',
    gap: space.md,
    minHeight: size.touch,
    paddingInline: space.sm,
    borderRadius: radius.md,
    color: color.textMuted,
    backgroundColor: { default: 'transparent', ':hover': color.bg },
    fontSize: '1.05rem',
    fontWeight: 700,
    textDecoration: 'none',
    transitionProperty: 'color, background-color',
    transitionDuration: { default: '150ms', [bp.reduceMotion]: '0s' },
    outlineStyle: { default: 'none', ':focus-visible': 'solid' },
    outlineWidth: 3,
    outlineColor: brand.yellow,
    outlineOffset: -3,
  },
  active: {
    color: color.link,
    backgroundColor: color.bg,
    boxShadow: `inset ${size.marker} 0 0 ${color.link}`,
  },
  slot: {
    display: 'flex',
    alignItems: 'center',
    justifyContent: 'center',
    flexShrink: 0,
    width: '2.25rem',
    height: '2.25rem',
  },
  icon: { width: '1.5rem', height: '1.5rem' },
});

type Props = Readonly<{ items: readonly NavItem[]; name: string; pictureUrl?: string }>;

// Menu lateral fixo (desktop): os mesmos itens do menu inferior, com rótulo ao lado do ícone.
export function SideNav({ items, name, pictureUrl }: Props) {
  const nav = useNavState(items);
  return (
    <nav aria-label={nav.title} {...stylex.props(styles.bar)}>
      <p {...stylex.props(styles.brand)}>{nav.appName}</p>
      <ul {...stylex.props(styles.list)}>
        {items.map((item) => {
          const active = nav.active(item);
          return (
            <li key={item.id}>
              <Link
                to={item.to}
                aria-current={active ? 'page' : undefined}
                {...stylex.props(styles.item, active && styles.active)}
              >
                <span {...stylex.props(styles.slot)}>
                  <NavGlyph item={item} name={name} pictureUrl={pictureUrl} {...stylex.props(styles.icon)} />
                </span>
                {nav.label(item)}
              </Link>
            </li>
          );
        })}
      </ul>
    </nav>
  );
}
