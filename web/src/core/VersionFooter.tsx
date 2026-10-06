import * as stylex from '@stylexjs/stylex';
import { useQuery } from '@tanstack/react-query';
import { useTranslation } from 'react-i18next';
import { getVersionOptions } from '../api/generated/@tanstack/react-query.gen';
import { color, space } from '../shared/ui/tokens/color.stylex';
import { webCommit, webVersion } from './version';

const styles = stylex.create({
  footer: { padding: space.md, color: color.text, opacity: 0.7, fontSize: '0.85rem', textAlign: 'center' },
});

// Rodape "v0.1.0 · API v0.1.0 · tst · a1b2c3d" (F-05). Sem resposta da API, mostra so a versao do web.
export function VersionFooter() {
  const { t } = useTranslation('common');
  const { data } = useQuery(getVersionOptions());
  return (
    <footer {...stylex.props(styles.footer)} data-testid="version-footer">
      {data
        ? t('footer.version', { web: webVersion, api: data.api, env: data.env, commit: webCommit })
        : t('footer.versionWebOnly', { web: webVersion, commit: webCommit })}
    </footer>
  );
}
