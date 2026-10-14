import * as stylex from '@stylexjs/stylex';
import { useTranslation } from 'react-i18next';
import { color, radius, space } from '../../../shared/ui/tokens/color.stylex';
import { loginUrl } from '../feature';
import { useSession } from '../hooks/useSession';
import { useSignOut } from '../hooks/useSignOut';

const styles = stylex.create({
  panel: { display: 'flex', flexDirection: 'column', alignItems: 'flex-start', gap: space.md, marginTop: space.lg },
  button: {
    display: 'inline-flex',
    alignItems: 'center',
    justifyContent: 'center',
    minHeight: '48px',
    minWidth: '48px',
    paddingInline: space.lg,
    borderRadius: radius.md,
    borderWidth: 0,
    backgroundColor: { default: color.primary, ':hover': color.primaryEdge },
    color: color.onPrimary,
    fontSize: '1rem',
    fontWeight: 600,
    textDecoration: 'none',
    cursor: 'pointer',
  },
  error: { color: color.danger, margin: 0 },
});

export function AuthPanel() {
  const { t, ready } = useTranslation('auth');
  const session = useSession();
  const { signOut, pending, failed } = useSignOut();
  if (!ready || session.status === 'loading') return <div aria-busy="true" />;
  if (session.status === 'anonymous') {
    return (
      <div {...stylex.props(styles.panel)}>
        <p>{t('greetingAnonymous')}</p>
        <a href={loginUrl} {...stylex.props(styles.button)}>
          {t('signIn')}
        </a>
      </div>
    );
  }
  return (
    <div {...stylex.props(styles.panel)}>
      <p>{t('greeting', { name: session.me.display_name })}</p>
      <button type="button" onClick={signOut} disabled={pending} {...stylex.props(styles.button)}>
        {t('signOut')}
      </button>
      {failed && (
        <p role="alert" {...stylex.props(styles.error)}>
          {t('signOutError')}
        </p>
      )}
    </div>
  );
}
