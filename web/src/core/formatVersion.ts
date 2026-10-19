// Formatacao do selo de versao: pura e sem textos de interface (os textos vem das chaves de i18n).
// A versao vem do CI (VERSION + numero de merges na develop, ex.: 1.0.162); em dev local e `dev`.

const SEMVER = /^\d+\.\d+\.\d+/;

export interface BadgeParts {
  /** Chave do texto em `common`: `footer.<kind>`. */
  kind: 'version' | 'versionWithApi';
  web: string;
  api?: string;
}

/** Versao do CI vira `v1.0.162`; fora do CI (dev local) o texto fica como veio, sem o `v`. */
const label = (version: string) => (SEMVER.test(version) ? `v${version}` : version);

/**
 * So a versao do web quando web e API estao iguais (mesmo commit); se a API esta em outra versao
 * (ex.: um merge so do web nao reconstroi a API), mostra as duas para ver que ambiente esta desatualizado.
 */
export function badgeParts(web: string, api?: string): BadgeParts {
  if (!api || label(api) === label(web)) return { kind: 'version', web: label(web) };
  return { kind: 'versionWithApi', web: label(web), api: label(api) };
}
