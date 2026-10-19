import * as stylex from '@stylexjs/stylex';
import { useQuery } from '@tanstack/react-query';
import { useTranslation } from 'react-i18next';
import { getMeOptions, getVersionOptions } from '../api/generated/@tanstack/react-query.gen';
import { color, radius, size, space } from '../shared/ui/tokens/color.stylex';
import { badgeParts } from './formatVersion';
import { webCommit, webVersion } from './version';

const styles = stylex.create({
  badge: {
    position: 'fixed',
    left: space.sm,
    bottom: space.sm,
    zIndex: 5,
    paddingBlock: space.xs,
    paddingInline: space.sm,
    borderRadius: radius.full,
    borderWidth: 1,
    borderStyle: 'solid',
    borderColor: color.border,
    backgroundColor: color.surface,
    color: color.textMuted,
    fontSize: '0.7rem',
    fontWeight: 700,
    fontVariantNumeric: 'tabular-nums',
    pointerEvents: 'none',
    userSelect: 'none',
  },
  aboveNav: { bottom: `calc(${size.nav} + ${space.sm})` },
});

// Selo de versao fixo no canto; sobe para ficar acima do menu inferior quando ha sessao.
export function VersionBadge() {
  const { t } = useTranslation('common');
  const version = useQuery(getVersionOptions());
  const me = useQuery({ ...getMeOptions(), retry: false, refetchOnWindowFocus: false });
  const api = version.data && { version: version.data.api, commit: version.data.commit, env: version.data.env };
  const { kind, ...values } = badgeParts({ version: webVersion, commit: webCommit }, api);
  return (
    <div {...stylex.props(styles.badge, me.data && styles.aboveNav)} data-testid="version-badge">
      {t(`footer.${kind}`, values)}
    </div>
  );
}
