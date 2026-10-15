import * as stylex from '@stylexjs/stylex';
import { useQuery } from '@tanstack/react-query';
import { useTranslation } from 'react-i18next';
import { getVersionOptions } from '../api/generated/@tanstack/react-query.gen';
import { color, space } from '../shared/ui/tokens/color.stylex';
import { footerParts } from './formatVersion';
import { webCommit, webVersion } from './version';

const styles = stylex.create({
  footer: { padding: space.md, color: color.text, opacity: 0.7, fontSize: '0.85rem', textAlign: 'center' },
});

export function VersionFooter() {
  const { t } = useTranslation('common');
  const { data } = useQuery(getVersionOptions());
  const web = { version: webVersion, commit: webCommit };
  const { kind, ...values } = footerParts(web, data && { version: data.api, commit: data.commit, env: data.env });
  return (
    <footer {...stylex.props(styles.footer)} data-testid="version-footer">
      {t(`footer.${kind}`, values)}
    </footer>
  );
}
