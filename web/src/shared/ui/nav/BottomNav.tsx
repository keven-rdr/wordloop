import * as stylex from '@stylexjs/stylex';
import { Link, useRouterState } from '@tanstack/react-router';
import { useTranslation } from 'react-i18next';
import { brand } from '../tokens/brand.stylex';
import { color, radius, size, space } from '../tokens/color.stylex';
import { bp } from '../tokens/media.stylex';
import { Avatar } from './Avatar';
import { CardsIcon, HomeIcon } from './Icons';

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
    maxWidth: '40rem',
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
      height: '3px',
      borderBottomLeftRadius: radius.full,
      borderBottomRightRadius: radius.full,
      backgroundColor: color.link,
    },
  },
  slot: { display: 'flex', alignItems: 'center', justifyContent: 'center', width: '2.25rem', height: '2.25rem' },
  icon: { width: '1.5rem', height: '1.5rem' },
});

type Props = Readonly<{ name: string; pictureUrl?: string }>;

const inside = (path: string, base: string) => path === base || path.startsWith(`${base}/`);

export function BottomNav({ name, pictureUrl }: Props) {
  const { t } = useTranslation('common');
  const path = useRouterState({ select: (state) => state.location.pathname });
  const onSettings = inside(path, '/settings') || inside(path, '/passkeys');
  const onWords = inside(path, '/words');
  return (
    <nav aria-label={t('nav.label')} {...stylex.props(styles.bar)}>
      <ul {...stylex.props(styles.list)}>
        <li {...stylex.props(styles.cell)}>
          <Link
            to="/settings"
            aria-current={onSettings ? 'page' : undefined}
            {...stylex.props(styles.item, onSettings && styles.active)}
          >
            <span {...stylex.props(styles.slot)}>
              <Avatar name={name} pictureUrl={pictureUrl} />
            </span>
            {t('nav.profile')}
          </Link>
        </li>
        <li {...stylex.props(styles.cell)}>
          <Link
            to="/"
            aria-current={path === '/' ? 'page' : undefined}
            {...stylex.props(styles.item, path === '/' && styles.active)}
          >
            <span {...stylex.props(styles.slot)}>
              <HomeIcon {...stylex.props(styles.icon)} />
            </span>
            {t('nav.home')}
          </Link>
        </li>
        <li {...stylex.props(styles.cell)}>
          <Link
            to="/words"
            aria-current={onWords ? 'page' : undefined}
            {...stylex.props(styles.item, onWords && styles.active)}
          >
            <span {...stylex.props(styles.slot)}>
              <CardsIcon {...stylex.props(styles.icon)} />
            </span>
            {t('nav.words')}
          </Link>
        </li>
      </ul>
    </nav>
  );
}
