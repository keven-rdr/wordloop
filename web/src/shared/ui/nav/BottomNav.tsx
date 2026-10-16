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
    backgroundColor: color.surface,
    borderTopWidth: 1,
    borderTopStyle: 'solid',
    borderTopColor: color.border,
    paddingBottom: 'env(safe-area-inset-bottom)',
  },
  list: {
    listStyle: 'none',
    margin: 0,
    padding: space.sm,
    display: 'flex',
    justifyContent: 'space-around',
    alignItems: 'center',
    gap: space.sm,
    maxWidth: '40rem',
    marginInline: 'auto',
    boxSizing: 'border-box',
  },
  item: {
    display: 'flex',
    flexDirection: 'column',
    alignItems: 'center',
    justifyContent: 'center',
    gap: space.xs,
    minWidth: size.touch,
    minHeight: size.touch,
    paddingInline: space.md,
    borderRadius: radius.md,
    color: color.textMuted,
    fontSize: '0.8rem',
    fontWeight: 700,
    textDecoration: 'none',
    transitionProperty: 'color, background-color',
    transitionDuration: { default: '150ms', [bp.reduceMotion]: '0s' },
    outlineStyle: { default: 'none', ':focus-visible': 'solid' },
    outlineWidth: 3,
    outlineColor: brand.yellow,
    outlineOffset: -3,
  },
  active: { color: color.link, backgroundColor: color.bg },
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
        <li>
          <Link
            to="/settings"
            aria-current={onSettings ? 'page' : undefined}
            {...stylex.props(styles.item, onSettings && styles.active)}
          >
            <Avatar name={name} pictureUrl={pictureUrl} />
            {t('nav.profile')}
          </Link>
        </li>
        <li>
          <Link
            to="/"
            aria-current={path === '/' ? 'page' : undefined}
            {...stylex.props(styles.item, path === '/' && styles.active)}
          >
            <HomeIcon {...stylex.props(styles.icon)} />
            {t('nav.home')}
          </Link>
        </li>
        <li>
          <Link
            to="/words"
            aria-current={onWords ? 'page' : undefined}
            {...stylex.props(styles.item, onWords && styles.active)}
          >
            <CardsIcon {...stylex.props(styles.icon)} />
            {t('nav.words')}
          </Link>
        </li>
      </ul>
    </nav>
  );
}
