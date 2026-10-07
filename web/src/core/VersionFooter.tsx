import * as stylex from '@stylexjs/stylex';
import { useQuery } from '@tanstack/react-query';
import { useTranslation } from 'react-i18next';
import { getVersionOptions } from '../api/generated/@tanstack/react-query.gen';
import { color, space } from '../shared/ui/tokens/color.stylex';
import { webCommit, webVersion } from './version';

const styles = stylex.create({
  footer: { padding: space.md, color: color.text, opacity: 0.7, fontSize: '0.85rem', textAlign: 'center' },
});

export function VersionFooter() {
  const { t } = useTranslation('common');
  const { data } = useQuery(getVersionOptions());
  const commit = data && data.commit !== webCommit ? `${webCommit}/${data.commit}` : webCommit;
  return (
    <footer {...stylex.props(styles.footer)} data-testid="version-footer">
      {data
        ? t('footer.version', { web: webVersion, api: data.api, env: data.env, commit })
        : t('footer.versionWebOnly', { web: webVersion, commit: webCommit })}
    </footer>
  );
}
