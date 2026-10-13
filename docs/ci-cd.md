# CI/CD: como está hoje

> Estado em 07/10/2026, na branch `feat/esqueleto-sprint-1`. O que já roda, o que é opcional e o que depende da VM e das contas. O desenho completo está em `docs/fase-4/04-engenharia.md`.

## Workflows (`.github/workflows/`)

| Workflow | Quando | O que faz | Check obrigatório |
|---|---|---|---|
| `api-ci` | PR e push em `develop` | `go generate`, `gofmt`, `go vet`, `golangci-lint` v2.14, `check-structure`, `go-arch-lint`, testes com `-race` e cobertura, `govulncheck`. No push com mudança na API: **imagem** `ghcr.io/<dono>/wordloop-api:sha-<commit>` (amd64 + arm64). Sonar opcional | `api-gate` |
| `web-ci` | PR e push em `develop` | `gen:api`, Biome, ESLint (StyleX), `tsc` 7, i18n, estrutura, Vitest com cobertura, **Playwright**. No push com mudança no web: imagem `wordloop-web:sha-<commit>` (amd64 + arm64). Sonar opcional | `web-gate` |
| `pr-title` | PR | Título em Conventional Commits (o merge é por squash: o título vira o commit) | `pr-title` |
| `security` | PR, push e toda segunda | CodeQL (Go e TypeScript), gitleaks no histórico, revisão de dependências novas (só PR) | não |
| `release` | push em `develop` | release-please por componente (`api-vX`, `web-vX`). **Desligado** até `RELEASE_PLEASE_ENABLED=true` | não |

Cada workflow filtra por caminho **dentro** dele; o job `*-gate` sempre roda e passa quando tudo foi OK **ou pulado**. Por isso a PR que só mexe em `docs/` não fica presa esperando check. Todas as actions de terceiros estão **fixadas por SHA** (conferido contra a tag).

## Configuração a fazer no GitHub (uma vez)

| O quê | Onde | Para quê |
|---|---|---|
| **Importar o ruleset** `.github/rulesets/develop.json` | Settings → Rules → Rulesets → New ruleset → *Import a ruleset* | Exige PR, merge por **squash**, `api-gate`, `web-gate` e `pr-title`; bloqueia exclusão e *force push* da `develop`. **Importe só depois do primeiro CI verde**, para conferir que os nomes dos checks batem: um nome errado como obrigatório trava toda PR |
| Permitir só squash e apagar branch ao mergear | Settings → General → Pull Requests | Mesma regra do ruleset |
| Permitir que Actions criem PRs | Settings → Actions → General → *Allow GitHub Actions to create and approve pull requests* | Necessário ao release-please |
| Instalar o app **Renovate** no repositório, **depois de mergear a PR do esqueleto** | github.com/apps/renovate → Install → *Select repositories* → `wordloop` | Lê o `renovate.json` **da branch padrão (`develop`)**. Instalado antes do merge, ele não acha a configuração e abre uma PR "Configure Renovate" com a configuração padrão dele |
| Conferir que o *Dependency graph* está ligado (em repositório público costuma vir ligado) | Settings → Code security (pode aparecer como *Advanced Security*) | A revisão de dependências do workflow `security` usa |
| **NÃO ligar o CodeQL *Default setup*** | Settings → Code security → Code scanning | O CodeQL já roda pelo workflow `security.yml` (*advanced setup*). Ligar o *default setup* **desativa esse workflow** e bloqueia os uploads dele ([documentação](https://docs.github.com/en/code-security/code-scanning/enabling-code-scanning/configuring-default-setup-for-code-scanning)). Os alertas aparecem sozinhos em Security → Code scanning |

### Opcionais (ligam sozinhos quando existirem)

| Recurso | Como ligar |
|---|---|
| **Sonar** | Crie os projetos `keven-rdr_wordloop-api` e `keven-rdr_wordloop-web` no SonarQube Cloud (plano gratuito; `develop` como branch principal). No repositório: secrets `SONAR_TOKEN_API` e `SONAR_TOKEN_WEB` e variável `SONAR_ENABLED=true`. Desative a *Automatic Analysis* no Sonar (a análise vem do CI, com cobertura) |
| **release-please** | Variável de repositório `RELEASE_PLEASE_ENABLED=true` (depois das permissões acima). Ele abre a PR de release a partir dos commits convencionais |

## Imagens

- Registro: **GHCR**, autenticado com o `GITHUB_TOKEN` do próprio workflow. O pacote nasce **privado**; para a VM puxar sem credencial, torne-o público em *Packages → Package settings* (o repositório é público).
- **Multi-arquitetura sem emulação:** a API cross-compila no runner (`TARGETARCH`) e o web só copia arquivos estáticos para a imagem final, então o mesmo build serve a VM ARM da Oracle e máquinas amd64.
- **Build uma vez:** a imagem `sha-<commit>` é a que será promovida (re-etiquetada) para a versão; nada é reconstruído por ambiente. O web não tem `build-arg` de ambiente: a configuração entra em runtime.
- O código gerado do contrato **não é versionado**: a API o gera dentro do `Dockerfile` e o web, também.

## Deploy de tst (VM Oracle Micro)

`deploy.yml` (manual) chama por SSH, com chave de **comando forçado**, o `deploy/scripts/deploy.sh` na VM, que faz `pull` das imagens `sha-<commit>` do GHCR e `up -d` (ou `rollback`). A VM roda Caddy + API + web; o Postgres é o Neon e **não há Keycloak** em tst ([ADR 0022](adr/0022-vm-oracle-micro-neon-sem-keycloak-em-tst.md)). Preparo da VM: `deploy/scripts/bootstrap-vm.sh`. Arquivos de tst em `deploy/tst/` (`secrets.env` e `versions.env` reais ficam só na VM).

## Ainda não existe

Deploy em prd, promoção por re-etiqueta com release-please, Keycloak fora do dev, backup. Os esqueletos estão em `docs/fase-4/`.

## Hooks locais (lefthook + commitlint)

Instalados por `npm install` na raiz (script `prepare`). `commit-msg`: Conventional Commits. `pre-commit`: `gofmt` nos `.go` e `biome ci` no web. `pre-push`: `npm run check` (só os componentes alterados). **Não use `--no-verify`.**
