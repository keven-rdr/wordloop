// Formatacao do selo de versao: pura e sem textos de interface (os textos vem das chaves de i18n).

const SEMVER = /^\d+\.\d+\.\d+/;
const SHORT = 7;
// Enquanto nao existe release (sem tag `web-v*`/`api-v*`), `git describe` devolve so o hash.
const NO_RELEASE = '0.0.0';

export interface Build {
  version: string;
  commit: string;
}

export interface BadgeParts {
  /** Chave do texto em `common`: `footer.<kind>`. */
  kind: 'version' | 'versionWebOnly';
  web: string;
  commit: string;
  api?: string;
  env?: string;
}

const shortCommit = (commit: string) => commit.slice(0, SHORT);
/** Release vira `v0.1.0`; sem release (a "versao" e so o hash) mostra `v0.0.0`, e o commit vem a parte. */
const label = ({ version }: Build) => `v${SEMVER.test(version) ? version : NO_RELEASE}`;

/** O commit curto do web e o completo da API sao o mesmo commit quando um e prefixo do outro. */
function commitLabel(web: string, api: string): string {
  const w = shortCommit(web);
  const a = shortCommit(api);
  return w === a ? w : `${w}/${a}`;
}

export function badgeParts(web: Build, api?: Build & { env: string }): BadgeParts {
  if (!api) return { kind: 'versionWebOnly', web: label(web), commit: shortCommit(web.commit) };
  return {
    kind: 'version',
    web: label(web),
    api: label(api),
    env: api.env,
    commit: commitLabel(web.commit, api.commit),
  };
}
