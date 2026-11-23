import * as stylex from '@stylexjs/stylex';
import { Link } from '@tanstack/react-router';
import { NavGlyph, useNavState } from './items';
import type { NavItem } from './types';

type Style = stylex.StyleXStyles;

// A aparência de cada menu (inferior ou lateral); `brand` mostra o nome do app acima dos itens.
export type NavLook = Readonly<{
  bar: Style;
  list: Style;
  cell?: Style;
  item: Style;
  active: Style;
  slot: Style;
  icon: Style;
  brand?: Style;
}>;

type Props = Readonly<{ items: readonly NavItem[]; name: string; pictureUrl?: string; look: NavLook }>;

// O menu principal: itens do registro das features, com o item da tela atual marcado. Os dois menus só mudam o `look`.
export function MainNav({ items, name, pictureUrl, look }: Props) {
  const nav = useNavState(items);
  return (
    <nav aria-label={nav.title} {...stylex.props(look.bar)}>
      {look.brand && <p {...stylex.props(look.brand)}>{nav.appName}</p>}
      <ul {...stylex.props(look.list)}>
        {items.map((item) => {
          const active = nav.active(item);
          return (
            <li key={item.id} {...stylex.props(look.cell)}>
              <Link
                to={item.to}
                aria-current={active ? 'page' : undefined}
                {...stylex.props(look.item, active && look.active)}
              >
                <span {...stylex.props(look.slot)}>
                  <NavGlyph item={item} name={name} pictureUrl={pictureUrl} {...stylex.props(look.icon)} />
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
