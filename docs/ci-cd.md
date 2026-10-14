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

### Sonar como condição de merge

O job `sonar` roda o scanner com `sonar.qualitygate.wait=true`: se o *quality gate* reprovar, o job falha e o `api-gate`/`web-gate` (checks obrigatórios do ruleset) falham junto. Com a variável **`SONAR_REQUIRED=true`**, o gate também falha se o Sonar **não rodou** (variável `SONAR_ENABLED` ou token ausentes). Sem `SONAR_REQUIRED`, o Sonar roda e avisa, mas um job pulado não bloqueia.

O gate é o padrão "Sonar way" do plano gratuito, aplicado ao **código novo** da PR: cobertura ≥ 80%, duplicação ≤ 3%, notas A em confiabilidade, segurança e manutenibilidade, e *security hotspots* revisados. PR só de documentação não tem código novo e passa.

**Ordem para ligar (importa):** 1) mergear a PR que adiciona a espera pelo gate; 2) abrir uma PR de teste e conferir que o job `sonar` fica verde; 3) só então criar a variável `SONAR_REQUIRED=true` (*Settings → Secrets and variables → Actions → Variables*). Ligada antes, toda PR falharia.

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
- A árvore de rotas do web (`routeTree.gen.ts`) também **não é versionada**: `npm run gen` a gera, o `web-ci` e o `Dockerfile` rodam esse passo antes do `tsc`/build, e o plugin do Vite a mantém em dev. Em dev, a API pode estar em outra porta: `API_PROXY_TARGET=http://localhost:18088 npm run dev:web` (padrão `http://localhost:8080`).

## Deploy de tst (VM Oracle Micro)

**Deploy 100% automático.** Todo push na `develop` que gera imagem nova faz o `api-ci`/`web-ci` chamar o `deploy.yml` (job `deploy-tst`). Ele monta um pacote com `compose.yaml`, `Caddyfile`, `api.env` e o `secrets.env` (com a `DATABASE_URL` do **secret** do GitHub) e o envia pela entrada padrão do SSH, com chave de **comando forçado**, ao `deploy/scripts/deploy.sh` da VM. O script grava os arquivos, faz `pull` das imagens `sha-<commit>` do GHCR, **migra o banco quando a API muda** (`api migrate`, nunca no rollback) e sobe com `up -d`. A VM roda Caddy + API + web; o Postgres é o Neon e **não há Keycloak** em tst ([ADR 0022](adr/0022-vm-oracle-micro-neon-sem-keycloak-em-tst.md)). Nada é gravado à mão na VM.

- **Secrets exigidos** (*Settings → Secrets and variables → Actions → Secrets*): `DEPLOY_SSH_KEY`, `DEPLOY_HOST`, `DEPLOY_KNOWN_HOSTS` e `DATABASE_URL` (connection string do Neon). Se faltar algum, o job falha dizendo qual. `DATABASE_URL` precisa ser **secret**, não *variable* (variável aparece em texto aberto).
- **Contas e convites:** o admin é criado pelo dono, no servidor, com `docker compose --env-file versions.env run --rm api create-admin --email voce@exemplo.com --name Voce` (em `/srv/wordloop/tst`, por SSH com a chave do dono; a chave de deploy é de comando forçado e não serve). O comando imprime um link de convite de uso único (15 min) montado com `APP_PUBLIC_URL` (já em `deploy/tst/api.env`). Ver [ADR 0028](adr/0028-contas-proprias-sem-senha-passkeys-depois-google-github.md): o app não guarda senha; o login por passkey entra na PR 2.
- **Passkeys (login sem senha):** `WEBAUTHN_RP_ID` e `WEBAUTHN_ORIGIN` (já em `deploy/tst/api.env`: `tst.wordloop.com.br` e `https://tst.wordloop.com.br`). Sem as duas, as rotas de passkey respondem 503 e a API sobe normalmente. Em dev local: `WEBAUTHN_RP_ID=localhost` e `WEBAUTHN_ORIGIN=http://localhost:5173` (a origem é a do navegador, a do Vite). A API limita `/api/v1/auth/*` a 30 pedidos por minuto por IP (rajada de 15) e corpos a 64 KB; o IP só vem do `X-Forwarded-For` (posto pelo Caddy) com `API_TRUST_PROXY=true`, já em `deploy/tst/api.env`; sem isso o cabeçalho é ignorado.
- **Login (opcional, tudo ou nada):** `OIDC_ISSUER`, `OIDC_CLIENT_ID`, `OIDC_CLIENT_SECRET` e `SESSION_KEY` (32 bytes em base64, `openssl rand -base64 32`). Sem eles a API sobe com o login desligado (`/auth/*` responde 503); com só alguns, o deploy falha dizendo o que falta. O callback de tst, `https://tst.wordloop.com.br/api/v1/auth/callback`, vem do `deploy/tst/api.env`.
- **Manual:** *Actions → deploy → Run workflow* escolhe as tags ou faz `rollback`.
- **Só configuração:** um push na `develop` que altere `deploy/tst/**` ou `deploy/scripts/**` (Caddyfile, compose, scripts) roda o `deploy.yml` sozinho, mantendo as versões atuais; o `deploy.sh` grava o `Caddyfile` no lugar e recarrega o Caddy sem derrubar conexões.
- **CSP:** o Caddy aplica uma política estrita ao app e à API. Só `/api/v1/docs/*` admite `unsafe-inline`, porque o Swagger UI embutido gera `<script>` e `<style>` inline; sem isso a página fica em branco.
- **Concorrência:** os dois CIs podem chamar o deploy ao mesmo tempo; o `deploy.yml` os serializa. Um push na `develop` nunca cancela o CI (e o deploy) em andamento; só PR cancela execução antiga.
- Preparo único da VM: `deploy/scripts/bootstrap-vm.sh`. Arquivos de tst em `deploy/tst/` (`versions.env` real fica só na VM).

## Ainda não existe

Deploy em prd, promoção por re-etiqueta com release-please, Keycloak fora do dev, backup. Os esqueletos estão em `docs/fase-4/`.

## Hooks locais (lefthook + commitlint)

Instalados por `npm install` na raiz (script `prepare`). `commit-msg`: Conventional Commits. `pre-commit`: `gofmt` nos `.go` e `biome ci` no web. `pre-push`: `npm run check` (só os componentes alterados). **Não use `--no-verify`.**

## Login em tst (IdP OIDC externo)

A VM de tst não comporta o Keycloak, então tst usa um provedor OIDC hospedado (qualquer um que siga o padrão; a API só precisa do *issuer*). Passo a passo, uma vez:

1. No provedor, crie uma aplicação **Web (confidencial)** com *Authorization Code* + PKCE.
2. Cadastre o *redirect URI* `https://tst.wordloop.com.br/api/v1/auth/callback` e o *logout/web origin* `https://tst.wordloop.com.br`.
3. Copie o *issuer* (a URL base que serve `/.well-known/openid-configuration`), o *client id* e o *client secret*.
4. No GitHub, em *Settings → Secrets and variables → Actions → Secrets*, crie `OIDC_ISSUER`, `OIDC_CLIENT_ID`, `OIDC_CLIENT_SECRET` e `SESSION_KEY`.
5. O próximo push na `develop` envia tudo à VM; confira `https://tst.wordloop.com.br/api/v1/auth/login` (deve redirecionar ao provedor).

Em dev o login usa o Keycloak do Compose (`deploy/dev`, realm `wordloop-dev`): defina `SESSION_KEY` no `api/.env` e entre com `aluno@wordloop.local` / `aluno`.
