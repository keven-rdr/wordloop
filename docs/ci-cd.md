# CI/CD: como está hoje

> Estado em 07/10/2026, na branch `feat/esqueleto-sprint-1`. O que já roda, o que é opcional e o que depende da VM e das contas. O desenho completo está em `docs/fase-4/04-engenharia.md`.

## Versão do app (automática, por componente)

Cada componente tem a sua contagem de versão, em `api/VERSION` e `web/VERSION` (`MAJOR.MINOR.PATCH`, hoje `1.0.0`). A cada merge na `develop` que muda o componente, o job `image` de `api-ci` e `web-ci` roda `scripts/bump-version.sh`: sobe o **patch** (+1), grava o número num commit `chore(release): atualizar versao de <componente> para X.Y.Z [skip ci]` na `develop` e usa esse número como `VERSION` da imagem. Merge que não toca o componente não sobe a versão dele. O app mostra as duas na tela **Configurações** (`v1.0.12 · API v1.0.31`), então dá para ver quando um ambiente está para trás. Para subir minor ou major, edite o arquivo `VERSION` na própria PR.

O commit de bump sai de um worktree separado (a imagem é construída do commit do merge) e repete até 5 vezes se outro bump gravar antes. O release-please foi removido: a versão não vem mais de tags.

**Configuração (uma vez), porque o bump comita direto na `develop`, que exige PR:**
1. Crie um *fine-grained personal access token* (seu usuário, só este repositório, permissão **Contents: Read and write**) e guarde como secret `RELEASE_PUSH_TOKEN` (Settings → Secrets and variables → Actions → Secrets).
2. No ruleset `develop` (Settings → Rules → Rulesets), em **Bypass list** adicione o papel **Repository admin** com modo *Always* (o `.github/rulesets/develop.json` já traz isso, para quem reimportar).
3. Crie a variável de repositório `VERSION_BUMP_ENABLED=true`. Sem ela (ou sem o token) o CI não sobe versão: usa o número que estiver no arquivo e nada quebra.

## Workflows (`.github/workflows/`)

| Workflow | Quando | O que faz | Check obrigatório |
|---|---|---|---|
| `api-ci` | PR e push em `develop` | `go generate`, `gofmt`, `go vet`, `golangci-lint` v2.14, `check-structure`, `go-arch-lint`, testes com `-race` e cobertura, `govulncheck`. No push com mudança na API: **imagem** `ghcr.io/<dono>/wordloop-api:sha-<commit>` (amd64 + arm64). Sonar opcional | `api-gate` |
| `web-ci` | PR e push em `develop` | `gen:api`, Biome, ESLint (StyleX), `tsc` 7, i18n, estrutura, Vitest com cobertura, **Playwright**. No push com mudança no web: imagem `wordloop-web:sha-<commit>` (amd64 + arm64). Sonar opcional | `web-gate` |
| `pr-title` | PR | Título em Conventional Commits (o merge é por squash: o título vira o commit) | `pr-title` |
| `security` | PR, push e toda segunda | CodeQL (Go e TypeScript), gitleaks no histórico, revisão de dependências novas (só PR) | não |

Cada workflow filtra por caminho **dentro** dele; o job `*-gate` sempre roda e passa quando tudo foi OK **ou pulado**. Por isso a PR que só mexe em `docs/` não fica presa esperando check. Todas as actions de terceiros estão **fixadas por SHA** (conferido contra a tag).

## Configuração a fazer no GitHub (uma vez)

| O quê | Onde | Para quê |
|---|---|---|
| **Importar o ruleset** `.github/rulesets/develop.json` | Settings → Rules → Rulesets → New ruleset → *Import a ruleset* | Exige PR, merge por **squash**, `api-gate`, `web-gate` e `pr-title`; bloqueia exclusão e *force push* da `develop`. **Importe só depois do primeiro CI verde**, para conferir que os nomes dos checks batem: um nome errado como obrigatório trava toda PR |
| Permitir só squash e apagar branch ao mergear | Settings → General → Pull Requests | Mesma regra do ruleset |
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
| **Bump de versão** | Secret `RELEASE_PUSH_TOKEN`, bypass do ruleset e variável `VERSION_BUMP_ENABLED=true` (ver a seção *Versão do app* acima) |

## Imagens

- Registro: **GHCR**, autenticado com o `GITHUB_TOKEN` do próprio workflow. O pacote nasce **privado**; para a VM puxar sem credencial, torne-o público em *Packages → Package settings* (o repositório é público).
- **Multi-arquitetura sem emulação:** a API cross-compila no runner (`TARGETARCH`) e o web só copia arquivos estáticos para a imagem final, então o mesmo build serve a VM ARM da Oracle e máquinas amd64.
- **Build uma vez:** a imagem `sha-<commit>` é a que será promovida (re-etiquetada) para a versão; nada é reconstruído por ambiente. O web não tem `build-arg` de ambiente: a configuração entra em runtime.
- O código gerado do contrato **não é versionado**: a API o gera dentro do `Dockerfile` e o web, também.
- A árvore de rotas do web (`routeTree.gen.ts`) também **não é versionada**: `npm run gen` a gera, o `web-ci` e o `Dockerfile` rodam esse passo antes do `tsc`/build, e o plugin do Vite a mantém em dev. Em dev, a API pode estar em outra porta: `API_PROXY_TARGET=http://localhost:18088 npm run dev:web` (padrão `http://localhost:8080`).

## Deploy de tst (VM Oracle Micro)

**Deploy 100% automático.** Todo push na `develop` que gera imagem nova faz o `api-ci`/`web-ci` chamar o `deploy.yml` (job `deploy-tst`). Ele monta um pacote com `compose.yaml`, `Caddyfile`, `api.env` e o `secrets.env` (com a `DATABASE_URL` do **secret** do GitHub) e o envia pela entrada padrão do SSH, com chave de **comando forçado**, ao `deploy/scripts/deploy.sh` da VM. O script grava os arquivos, faz `pull` das imagens `sha-<commit>` do GHCR, **migra o banco quando a API muda** (`api migrate`, nunca no rollback) e sobe com `up -d`. A VM roda Caddy + API + web; o Postgres é o Neon e **não há Keycloak** em tst ([ADR 0022](adr/0022-vm-oracle-micro-neon-sem-keycloak-em-tst.md)). Nada é gravado à mão na VM.

- **Secrets exigidos** (*Settings → Secrets and variables → Actions → Secrets*): `DEPLOY_SSH_KEY`, `DEPLOY_HOST`, `DEPLOY_KNOWN_HOSTS` e `DATABASE_URL` (connection string do Neon). Se faltar algum, o job falha dizendo qual. `DATABASE_URL` precisa ser **secret**, não *variable* (variável aparece em texto aberto).
- **Contas e convites:** o admin é criado pelo dono, no servidor, com `docker compose --env-file versions.env run --rm api create-admin --email voce@exemplo.com --name Voce` (em `/srv/wordloop/tst`, por SSH com a chave do dono; a chave de deploy é de comando forçado e não serve). O comando imprime um link de convite de uso único (15 min) montado com `APP_PUBLIC_URL` (já em `deploy/tst/api.env`). Ver [ADR 0028](adr/0028-contas-proprias-sem-senha-passkeys-depois-google-github.md): o app não guarda senha. **Sem acesso SSH** (ou para criar o admin direto no banco), rode `deploy/scripts/create-admin.sql` no *SQL Editor* do Neon: troque o e-mail e o nome no início do arquivo e a resposta é o link do convite (o banco guarda só o hash, como o comando); o link abre a tela em que o admin **define a própria senha** (ou cria uma passkey). Esqueceu a senha: `deploy/scripts/reset-admin-access.sql` gera outro convite, apaga as sessões e destrava a conta. A senha vira hash Argon2id ([ADR 0029](adr/0029-senha-como-hash-argon2id-ao-lado-da-passkey.md)); nunca passa por SQL.
- **E-mail (cadastro e recuperação de senha):** conta no Brevo (plano gratuito), o domínio `wordloop.com.br` verificado como remetente (registros DNS de DKIM/SPF que o Brevo mostra) e o secret `BREVO_API_KEY` no GitHub (*Settings → Secrets and variables → Actions → Secrets*). O remetente (`MAIL_FROM` e `MAIL_FROM_NAME`) está em `deploy/tst/api.env`; troque se usar outro endereço verificado. Sem a chave, `POST /auth/register` e `POST /auth/password/forgot` respondem 503 e a tela mostra que o envio não está disponível. Em desenvolvimento, `MAIL_DRIVER=log` escreve o e-mail (com o link) no log da API, sem enviar nada. Os links vão com o segredo depois do `#`, que o navegador não manda ao servidor.
- **Passkeys (login sem senha):** `WEBAUTHN_RP_ID` e `WEBAUTHN_ORIGIN` (já em `deploy/tst/api.env`: `tst.wordloop.com.br` e `https://tst.wordloop.com.br`). Sem as duas, as rotas de passkey respondem 503 e a API sobe normalmente. Em dev local: `WEBAUTHN_RP_ID=localhost` e `WEBAUTHN_ORIGIN=http://localhost:5173` (a origem é a do navegador, a do Vite). A API limita `/api/v1/auth/*` a 30 pedidos por minuto por IP (rajada de 15) e corpos a 64 KB; o IP só vem do `X-Forwarded-For` (posto pelo Caddy) com `API_TRUST_PROXY=true`, já em `deploy/tst/api.env`; sem isso o cabeçalho é ignorado.
- **Login social (opcional):** `GOOGLE_CLIENT_ID`/`GOOGLE_CLIENT_SECRET` e `GH_OAUTH_CLIENT_ID`/`GH_OAUTH_CLIENT_SECRET` (o GitHub não aceita nome de secret começando com `GITHUB_`; o deploy os entrega à API como `GITHUB_CLIENT_ID/SECRET`), mais `SESSION_KEY` (32 bytes em base64, `openssl rand -base64 32`), obrigatório com qualquer provedor. Cada par é tudo ou nada; sem nenhum, só e-mail e senha e passkey funcionam e os botões ficam como "em breve". `OIDC_ISSUER`/`OIDC_CLIENT_ID`/`OIDC_CLIENT_SECRET` ligam um provedor OIDC genérico (usado em dev com o Keycloak). O callback, o mesmo para todos, vem de `OIDC_REDIRECT_URL` em `deploy/tst/api.env`. Passo a passo em [Login social em tst](#login-social-em-tst).
- **Manual:** *Actions → deploy → Run workflow* escolhe as tags ou faz `rollback`.
- **Smoke test:** depois do deploy, o último passo espera a API ficar pronta (até 2 min) e confere que o commit servido em `/api/v1/version` é o da tag implantada, que `/` devolve a página do app e que `/favicon.svg` e `/api/v1/docs/` respondem. Se algo falhar, o job vermelho aponta o que foi. Ele **não desfaz** o deploy sozinho (migração não se desfaz): quem decide é o rollback manual.
- **Só configuração:** um push na `develop` que altere `deploy/tst/**` ou `deploy/scripts/**` (Caddyfile, compose, scripts) roda o `deploy.yml` sozinho, mantendo as versões atuais; o `deploy.sh` grava o `Caddyfile` no lugar e recarrega o Caddy sem derrubar conexões.
- **CSP:** o Caddy aplica uma política estrita ao app e à API. Só `/api/v1/docs/*` admite `unsafe-inline`, porque o Swagger UI embutido gera `<script>` e `<style>` inline; sem isso a página fica em branco.
- **Concorrência:** os dois CIs podem chamar o deploy ao mesmo tempo; o `deploy.yml` os serializa. Um push na `develop` nunca cancela o CI (e o deploy) em andamento; só PR cancela execução antiga.
- Preparo único da VM: `deploy/scripts/bootstrap-vm.sh`. Arquivos de tst em `deploy/tst/` (`versions.env` real fica só na VM).

## Ainda não existe

Deploy em prd, promoção por re-etiqueta da imagem, Keycloak fora do dev, backup. Os esqueletos estão em `docs/fase-4/`.

## Hooks locais (lefthook + commitlint)

Instalados por `npm install` na raiz (script `prepare`). `commit-msg`: Conventional Commits. `pre-commit`: `gofmt` nos `.go` e `biome ci` no web. `pre-push`: `npm run check` (só os componentes alterados). **Não use `--no-verify`.**

## Login social em tst

O login com Google e GitHub **vincula ou cria** a conta pelo **e-mail confirmado** que o provedor informa: se já existe conta com aquele e-mail (de senha, passkey ou convite), a pessoa entra nela e o vínculo fica guardado; se não existe, a conta é criada já com o e-mail confirmado. Sem e-mail confirmado no provedor, o login é recusado (`/login?error=email`). Quem ligou o 2FA na conta não entra só pelo provedor (`/login?error=two_factor`).

Passo a passo, uma vez por provedor (o endereço de retorno é o mesmo nos dois: `https://tst.wordloop.com.br/api/v1/auth/callback`):

**Google**
1. [Google Cloud Console](https://console.cloud.google.com/) → crie um projeto → *APIs & Services → OAuth consent screen*: tipo *External*, nome `wordloop`, seu e-mail; escopos `openid`, `email` e `profile`.
2. *Credentials → Create credentials → OAuth client ID* → *Web application*. Em *Authorized redirect URIs* coloque o endereço de retorno acima. Copie o *Client ID* e o *Client secret*.
3. No GitHub (*Settings → Secrets and variables → Actions → Secrets*): `GOOGLE_CLIENT_ID` e `GOOGLE_CLIENT_SECRET`.

**GitHub**
1. GitHub → *Settings → Developer settings → OAuth Apps → New OAuth App*. *Homepage URL* `https://tst.wordloop.com.br` e *Authorization callback URL* o endereço de retorno acima.
2. Gere um *Client secret* e copie com o *Client ID*.
3. Secrets `GH_OAUTH_CLIENT_ID` e `GH_OAUTH_CLIENT_SECRET`.

**Os dois:** crie também o secret `SESSION_KEY` (rode `openssl rand -base64 32` no terminal e cole o resultado) e faça um push na `develop` (ou rode o workflow *deploy*). Confira em `https://tst.wordloop.com.br/api/v1/auth/providers`: deve listar `google` e/ou `github`; os botões da tela de login passam a funcionar sozinhos. **Nunca mande o client secret ou a `SESSION_KEY` no chat.**

Em dev o login social usa o Keycloak do Compose (`deploy/dev`, realm `wordloop-dev`, provedor `oidc`): defina `SESSION_KEY` no `api/.env` e entre com `aluno@wordloop.local` / `aluno`.
