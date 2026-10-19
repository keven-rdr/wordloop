import * as stylex from '@stylexjs/stylex';
import { useQuery } from '@tanstack/react-query';
import { useTranslation } from 'react-i18next';
import { getVersionOptions } from '../api/generated/@tanstack/react-query.gen';
import { color, radius, space } from '../shared/ui/tokens/color.stylex';
import { badgeParts } from './formatVersion';
import { webVersion } from './version';

const styles = stylex.create({
  badge: {
    alignSelf: 'flex-start',
    paddingBlock: space.xs,
    paddingInline: space.sm,
    borderRadius: radius.full,
    borderWidth: 1,
    borderStyle: 'solid',
    borderColor: color.border,
    backgroundColor: color.surface,
    color: color.textMuted,
    fontSize: '0.75rem',
    fontWeight: 700,
    fontVariantNumeric: 'tabular-nums',
    userSelect: 'none',
  },
});

// Selo da versao (web e, se diferente, API); aparece na tela de Configuracoes.
export function VersionBadge() {
  const { t } = useTranslation('common');
  const version = useQuery(getVersionOptions());
  const { kind, ...values } = badgeParts(webVersion, version.data?.api);
  return (
    <div {...stylex.props(styles.badge)} data-testid="version-badge">
      {t(`footer.${kind}`, values)}
    </div>
  );
}
