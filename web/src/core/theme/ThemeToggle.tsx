import * as stylex from '@stylexjs/stylex';
import { useTranslation } from 'react-i18next';
import { color, radius, size } from '../../shared/ui/tokens/color.stylex';
import { effectiveTheme, setThemeChoice, useThemeChoice } from './useTheme';

const styles = stylex.create({
  button: {
    display: 'inline-flex',
    alignItems: 'center',
    justifyContent: 'center',
    width: size.touch,
    height: size.touch,
    borderRadius: radius.full,
    borderWidth: 1,
    borderStyle: 'solid',
    borderColor: color.border,
    backgroundColor: { default: 'transparent', ':hover': color.surface },
    color: color.text,
    cursor: 'pointer',
    outlineStyle: { default: 'none', ':focus-visible': 'solid' },
    outlineWidth: 2,
    outlineColor: color.link,
    outlineOffset: 2,
  },
  icon: { width: '1.4rem', height: '1.4rem' },
});

export function ThemeToggle() {
  const { t } = useTranslation('common');
  const current = effectiveTheme(useThemeChoice());
  const next = current === 'dark' ? 'light' : 'dark';
  return (
    <button
      type="button"
      onClick={() => setThemeChoice(next)}
      aria-label={t(next === 'light' ? 'theme.toLight' : 'theme.toDark')}
      {...stylex.props(styles.button)}
    >
      <svg
        viewBox="0 0 24 24"
        fill="none"
        stroke="currentColor"
        strokeWidth="2"
        strokeLinecap="round"
        strokeLinejoin="round"
        aria-hidden="true"
        {...stylex.props(styles.icon)}
      >
        {current === 'dark' ? (
          <>
            <circle cx="12" cy="12" r="4" />
            <path d="M12 2v2M12 20v2M4.9 4.9l1.4 1.4M17.7 17.7l1.4 1.4M2 12h2M20 12h2M4.9 19.1l1.4-1.4M17.7 6.3l1.4-1.4" />
          </>
        ) : (
          <path d="M21 12.8A9 9 0 1 1 11.2 3a7 7 0 0 0 9.8 9.8z" />
        )}
      </svg>
    </button>
  );
}
