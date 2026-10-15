// Formatacao do rodape de versao: pura e sem textos de interface (os textos vem das chaves de i18n).

const SEMVER = /^\d+\.\d+\.\d+/;
const SHORT = 7;

export interface Build {
  version: string;
  commit: string;
}

export interface FooterParts {
  /** Chave do texto em `common`: `footer.<kind>`. */
  kind: 'version' | 'versionWebOnly' | 'build' | 'buildWebOnly';
  web: string;
  api?: string;
  env?: string;
  commit?: string;
}

const shortCommit = (commit: string) => commit.slice(0, SHORT);
const isRelease = (version: string) => SEMVER.test(version);
/** Versao de release vira `v0.1.0`; num build comum a "versao" e so o hash e mostramos o commit curto. */
const label = ({ version, commit }: Build) => (isRelease(version) ? `v${version}` : shortCommit(commit));

/** O commit curto do web e o completo da API sao o mesmo commit quando um e prefixo do outro. */
function commitLabel(web: string, api: string): string {
  const w = shortCommit(web);
  const a = shortCommit(api);
  return w === a ? w : `${w}/${a}`;
}

export function footerParts(web: Build, api?: Build & { env: string }): FooterParts {
  if (!api) {
    return isRelease(web.version)
      ? { kind: 'versionWebOnly', web: label(web), commit: shortCommit(web.commit) }
      : { kind: 'buildWebOnly', web: label(web) };
  }
  if (isRelease(web.version) || isRelease(api.version)) {
    return {
      kind: 'version',
      web: label(web),
      api: label(api),
      env: api.env,
      commit: commitLabel(web.commit, api.commit),
    };
  }
  return { kind: 'build', web: label(web), api: label(api), env: api.env };
}
