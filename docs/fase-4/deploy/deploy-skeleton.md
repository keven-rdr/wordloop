# Esqueleto de deploy (Fase 4)

> Rascunho de arquivos que viverão em `deploy/`. **Não foram executados.** Versões de imagem: sempre por *tag + digest* fixos (nunca `latest`); os `...` indicam o digest a preencher no Sprint 0. Itens `[VERIFICAR]` entram nos spikes.

## 1. Layout na VM

```
/srv/wordloop/
├── shared/        compose.yaml  Caddyfile  .env (600)     # Caddy, PostgreSQL, Keycloak (UM Keycloak, dois realms)
├── tst/           compose.yaml  .env.versions  .env.secrets  environment.json
├── prd/           compose.yaml  .env.versions  .env.secrets  environment.json
├── media/{tst,prd}/                                       # imagens (fora do Git)
├── backups/       backup.sh  (pg_dump criptografado → fora da VM)
└── deploy.sh      # comando FORÇADO da chave SSH de deploy
```
Projetos Compose separados (`wl-shared`, `wl-tst`, `wl-prd`) e redes `wl-edge` (Caddy ↔ web/api/keycloak) e `wl-data` (apps ↔ Postgres). **Só o Caddy publica portas** (80/443): porta publicada pelo Docker contorna o `ufw`, então Postgres e Keycloak **não** publicam nada.

## 2. `shared/compose.yaml`

```yaml
name: wl-shared
networks:
  edge: { name: wl-edge }
  data: { name: wl-data }
volumes: { caddy_data: {}, caddy_config: {}, pgdata: {} }

services:
  caddy:
    image: caddy:2@sha256:...
    restart: unless-stopped
    ports: ["80:80", "443:443", "443:443/udp"]
    volumes:
      - ./Caddyfile:/etc/caddy/Caddyfile:ro
      - caddy_data:/data
      - caddy_config:/config
      - /srv/wordloop/media:/srv/media:ro
    networks: [edge]
    mem_limit: 128m

  postgres:                                  # UMA instância, três bancos: wordloop_tst, wordloop_prd, keycloak (um papel por banco)
    image: postgres:17-alpine@sha256:...     # [VERIFICAR] versão estável atual
    restart: unless-stopped
    environment: { POSTGRES_PASSWORD_FILE: /run/secrets/pg_admin }
    secrets: [pg_admin]
    volumes: [pgdata:/var/lib/postgresql/data, ./initdb:/docker-entrypoint-initdb.d:ro]
    networks: [data]
    mem_limit: 1g
    healthcheck: { test: ["CMD-SHELL", "pg_isready -U postgres"], interval: 10s, retries: 5 }

  keycloak:
    image: quay.io/keycloak/keycloak:26.7.1@sha256:...
    restart: unless-stopped
    command: ["start", "--import-realm"]     # importa SÓ realms inexistentes; segredos entram por placeholder de ambiente
    environment:
      KC_DB: postgres
      KC_DB_URL: jdbc:postgresql://postgres:5432/keycloak
      KC_DB_USERNAME: keycloak
      KC_DB_PASSWORD_FILE: /run/secrets/kc_db     # [VERIFICAR] suporte a *_FILE nesta versão; senão variável de ambiente
      KC_HOSTNAME: https://id.wordloop.com.br
      KC_HTTP_ENABLED: "true"                    # TLS termina no Caddy
      KC_PROXY_HEADERS: xforwarded               # [VERIFICAR] nome da opção na 26.x
      KC_HEALTH_ENABLED: "true"
      JAVA_OPTS_KC_HEAP: "-Xms256m -Xmx768m"      # medir no spike; alvo ~1,25 GB de contêiner
    volumes: [./realms:/opt/keycloak/data/import:ro]
    networks: [edge, data]
    mem_limit: 1536m
    depends_on: { postgres: { condition: service_healthy } }

secrets:
  pg_admin: { file: ./secrets/pg_admin }
  kc_db:    { file: ./secrets/kc_db }
```
**Realms como código:** `realms/wordloop-tst.json` e `wordloop-prd.json` (sem segredos, com placeholders). Console de administração do Keycloak **não** é exposto à internet (ver Caddyfile): acesso por túnel SSH.

## 3. `shared/Caddyfile`

```caddy
{
	email voce@exemplo.com            # contato do Let's Encrypt
}

(app) {
	encode zstd gzip
	header {
		Strict-Transport-Security "max-age=31536000"
		X-Content-Type-Options nosniff
		Referrer-Policy strict-origin-when-cross-origin
		# CSP: StyleX dinâmico usa atributo style="--var:..."; precisa style-src-attr [VERIFICAR no spike]
		Content-Security-Policy "default-src 'self'; script-src 'self'; style-src 'self'; style-src-attr 'unsafe-inline'; img-src 'self' data:; connect-src 'self'; frame-ancestors 'none'"
	}
}

wordloop.com.br {
	import app
	handle /api/*   { reverse_proxy wl-prd-api:8080 }
	handle /media/* { root * /srv/media/prd ; uri strip_prefix /media ; file_server }
	handle          { reverse_proxy wl-prd-web:8080 }
}

tst.wordloop.com.br {
	import app
	handle /api/*   { reverse_proxy wl-tst-api:8080 }
	handle /media/* { root * /srv/media/tst ; uri strip_prefix /media ; file_server }
	handle          { reverse_proxy wl-tst-web:8080 }
}

id.wordloop.com.br {
	encode zstd gzip
	@admin path /admin/* /realms/master/*
	respond @admin 403                  # administração só por túnel SSH
	reverse_proxy keycloak:8080
}
```

## 4. `base/compose.yaml` (um conjunto por ambiente; `ENV=tst|prd`)

```yaml
name: wl-${ENV}
networks:
  edge: { external: true, name: wl-edge }
  data: { external: true, name: wl-data }

x-hardening: &hardening
  restart: unless-stopped
  read_only: true
  cap_drop: [ALL]
  security_opt: ["no-new-privileges:true"]

services:
  api:
    <<: *hardening
    container_name: wl-${ENV}-api
    image: ghcr.io/OWNER/wordloop-api:${API_TAG}
    command: ["/api", "serve"]
    environment: { ROLES: http, ENV: "${ENV}", OIDC_ISSUER: "https://id.wordloop.com.br/realms/wordloop-${ENV}" }
    env_file: [.env.secrets]                 # DATABASE_URL, OIDC_CLIENT_SECRET, SESSION_KEY, VAPID_*
    networks: [edge, data]
    mem_limit: 256m
    healthcheck: { test: ["CMD", "/api", "healthcheck"], interval: 15s, retries: 5 }

  worker:
    <<: *hardening
    container_name: wl-${ENV}-worker
    image: ghcr.io/OWNER/wordloop-api:${API_TAG}     # MESMA imagem da api
    command: ["/api", "serve"]
    environment: { ROLES: worker, ENV: "${ENV}" }
    env_file: [.env.secrets]
    networks: [data]
    mem_limit: 192m

  web:
    <<: *hardening
    container_name: wl-${ENV}-web
    image: ghcr.io/OWNER/wordloop-web:${WEB_TAG}
    volumes: [./environment.json:/usr/share/nginx/html/environment.json:ro]   # config em RUNTIME
    tmpfs: [/tmp, /var/cache/nginx]
    networks: [edge]
    mem_limit: 64m
```
**Sobreposições por ambiente** (`overlays/tst`, `overlays/prd`): `.env.versions` (tags), `environment.json` (URL pública, nome do ambiente), limites de memória e domínio. Segredos **nunca** no Git: `.env.secrets` é criado na VM (modo 600); o repositório só tem `.env.secrets.example`. **Credenciais de dev não servem em tst/prd** (valores gerados por ambiente).

## 5. `web/nginx.conf` (cache correto)

```nginx
server {
  listen 8080;
  root /usr/share/nginx/html;
  gzip on;
  gzip_types text/css application/javascript application/json image/svg+xml;

  location = /health { default_type text/plain; return 200 "ok"; }

  location /assets/ {                       # arquivos com hash: imutáveis
    add_header Cache-Control "public, max-age=31536000, immutable";
    try_files $uri =404;
  }
  location = /sw.js                { add_header Cache-Control "no-cache"; try_files $uri =404; }
  location = /manifest.webmanifest { add_header Cache-Control "no-cache"; try_files $uri =404; }
  location = /environment.json     { add_header Cache-Control "no-store"; try_files $uri =404; }
  location = /index.html           { add_header Cache-Control "no-cache"; }

  location / {                              # fallback de SPA
    add_header Cache-Control "no-cache";
    try_files $uri /index.html;
  }
}
```
`sw.js`, `index.html`, manifest e `environment.json` **nunca** com cache longo (no Contratos o nginx marca todo `.js` como imutável por 1 ano, inclusive um `sw.js` na raiz).

## 6. Dockerfiles

```dockerfile
# web/Dockerfile — SEM build-arg de ambiente
FROM node:24-alpine AS build            # [VERIFICAR] LTS vigente
ARG VERSION=dev
ARG COMMIT=unknown
WORKDIR /app
COPY package.json package-lock.json ./
RUN npm ci
COPY . .
ENV VITE_APP_VERSION=$VERSION VITE_APP_COMMIT=$COMMIT
RUN npm run build

FROM nginxinc/nginx-unprivileged:1.28-alpine
COPY nginx.conf /etc/nginx/conf.d/default.conf
COPY --from=build /app/dist /usr/share/nginx/html
```
```dockerfile
# api/Dockerfile
FROM golang:1.27-alpine AS build
ARG VERSION=dev
ARG COMMIT=unknown
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -trimpath \
    -ldflags "-s -w -X wordloop/internal/platform/version.Version=${VERSION} -X wordloop/internal/platform/version.Commit=${COMMIT}" \
    -o /out/api ./cmd/api

FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=build /out/api /api
COPY --from=build /src/migrations /migrations
ENTRYPOINT ["/api"]
```

## 7. `deploy.sh` (comando forçado)

`~deploy/.ssh/authorized_keys`: `command="/srv/wordloop/deploy.sh",no-port-forwarding,no-agent-forwarding,no-X11-forwarding,no-pty ssh-ed25519 AAAA...`

```bash
#!/usr/bin/env bash
set -euo pipefail
read -r ENV API_TAG WEB_TAG <<<"${SSH_ORIGINAL_COMMAND:-}"      # '-' = manter a tag atual
[[ "$ENV" =~ ^(tst|prd)$ ]] || { echo "env inválido" >&2; exit 2; }
tag_ok() { [[ "$1" == "-" || "$1" =~ ^[A-Za-z0-9._-]{1,128}$ ]]; }
tag_ok "$API_TAG" && tag_ok "$WEB_TAG" || { echo "tag inválida" >&2; exit 2; }

cd "/srv/wordloop/$ENV"
cp .env.versions ".env.versions.prev"                            # base do rollback
[[ "$API_TAG" == "-" ]] || sed -i "s|^API_TAG=.*|API_TAG=$API_TAG|" .env.versions
[[ "$WEB_TAG" == "-" ]] || sed -i "s|^WEB_TAG=.*|WEB_TAG=$WEB_TAG|" .env.versions

export ENV; set -a; . ./.env.versions; set +a
docker compose --env-file .env.versions pull
docker compose --env-file .env.versions run --rm api /api migrate up     # migrações ANTES (expand/contract)
docker compose --env-file .env.versions up -d --remove-orphans
```
**Migrações compatíveis com rollback:** *expand/contract*: uma versão só **adiciona** (coluna nova, tabela nova, nullable); a remoção vem em versão posterior, quando a anterior já não roda. Rollback = rodar o `deploy` de novo com a tag anterior; **não** se desfaz migração em produção.

## 8. `backups/backup.sh` (esboço)

`pg_dump -Fc` por banco (`wordloop_prd`, `wordloop_tst`, `keycloak`) → `age` (chave pública; privada fora da VM) → `rclone` para o bucket (Backblaze B2 ou Cloudflare R2 `[VERIFICAR]` camada gratuita) → *ping* no healthchecks.io (alerta se faltar). Retenção 7 diários + 4 semanais. **Teste de restauração mensal** restaurando o dump do prd em um banco descartável. Exportar também os realms (`kc.sh export`) após mudança de configuração e antes de atualizar o Keycloak.

## 9. Dev local (`deploy/dev/compose.yaml`, Windows + Docker Desktop)

`postgres` (imagem fixa, credenciais **só de dev**), `mailpit` (SMTP de teste, `:8025`), `keycloak` com `--import-realm` do realm `wordloop-dev` (usuários de teste no próprio JSON) e **perfil opcional `observability`** (Prometheus, Loki, Alloy, Grafana). Sem RabbitMQ, Meilisearch ou MinIO. `api` e `web` rodam fora do Compose (`go run`, `vite`) para recarga rápida.
