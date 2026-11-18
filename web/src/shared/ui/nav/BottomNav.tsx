import * as stylex from '@stylexjs/stylex';
import { Link, useRouterState } from '@tanstack/react-router';
import type { ReactNode } from 'react';
import { useTranslation } from 'react-i18next';
import { brand } from '../tokens/brand.stylex';
import { color, radius, size, space } from '../tokens/color.stylex';
import { bp } from '../tokens/media.stylex';
import { Avatar } from './Avatar';
import { CardsIcon, HomeIcon, StudyIcon } from './Icons';

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

type TabProps = Readonly<{
  to: '/' | '/settings' | '/study' | '/words';
  active: boolean;
  label: string;
  children: ReactNode;
}>;

function NavTab({ to, active, label, children }: TabProps) {
  return (
    <li {...stylex.props(styles.cell)}>
      <Link to={to} aria-current={active ? 'page' : undefined} {...stylex.props(styles.item, active && styles.active)}>
        <span {...stylex.props(styles.slot)}>{children}</span>
        {label}
      </Link>
    </li>
  );
}

export function BottomNav({ name, pictureUrl }: Props) {
  const { t } = useTranslation('common');
  const path = useRouterState({ select: (state) => state.location.pathname });
  const onSettings = inside(path, '/settings') || inside(path, '/passkeys');
  return (
    <nav aria-label={t('nav.label')} {...stylex.props(styles.bar)}>
      <ul {...stylex.props(styles.list)}>
        <NavTab to="/settings" active={onSettings} label={t('nav.profile')}>
          <Avatar name={name} pictureUrl={pictureUrl} />
        </NavTab>
        <NavTab to="/" active={path === '/'} label={t('nav.home')}>
          <HomeIcon {...stylex.props(styles.icon)} />
        </NavTab>
        <NavTab to="/study" active={inside(path, '/study')} label={t('nav.study')}>
          <StudyIcon {...stylex.props(styles.icon)} />
        </NavTab>
        <NavTab to="/words" active={inside(path, '/words')} label={t('nav.words')}>
          <CardsIcon {...stylex.props(styles.icon)} />
        </NavTab>
      </ul>
    </nav>
  );
}
