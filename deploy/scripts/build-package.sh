#!/usr/bin/env bash
# Monta o pacote de configuracao que o workflow de deploy envia por SSH ao deploy.sh da VM.
#   build-package.sh <tst|prd> <saida.tar>
# Segredos entram por variavel de ambiente (vindos dos secrets do GitHub) e SO existem dentro do pacote (secrets.env).
#   obrigatorio: DATABASE_URL
#   login (cada grupo e tudo ou nada): OIDC_ISSUER+OIDC_CLIENT_ID+OIDC_CLIENT_SECRET, GOOGLE_CLIENT_ID+_SECRET,
#   GITHUB_CLIENT_ID+_SECRET; com qualquer um, SESSION_KEY e obrigatorio. E-mail: BREVO_API_KEY (opcional).
#   prd: PRD_DOMAIN (obrigatorio) e MAIL_FROM (padrao nao-responda@PRD_DOMAIN) preenchem o api.env.tmpl.
set -euo pipefail

here="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
. "$here/lib/common.sh"

env_name="${1:-}"
out="${2:-}"
[[ "$env_name" == tst || "$env_name" == prd ]] || die "uso: build-package.sh <tst|prd> <saida.tar>" 2
[[ -n "$out" ]] || die "uso: build-package.sh <tst|prd> <saida.tar>" 2
src="${PACKAGE_SRC:-$(cd "$here/../.." && pwd)}/deploy/$env_name"
need_env DATABASE_URL

strip() { printf '%s' "$1" | tr -d '\r\n'; }

group_ok() { # tudo ou nada: 0 se o grupo esta completo, 1 se vazio; erro se pela metade
  local total=0 set=0 v
  for v in "$@"; do
    total=$((total + 1))
    if [[ -n "${!v:-}" ]]; then set=$((set + 1)); fi
  done
  if [[ "$set" -ne 0 && "$set" -ne "$total" ]]; then die "grupo incompleto: defina todos ou nenhum de: $*" 2; fi
  [[ "$set" -eq "$total" ]]
}

emit() { local v; for v in "$@"; do printf '%s=%s\n' "$v" "$(strip "${!v}")"; done; }

render_api_env() { # prd: preenche o modelo; tst: copia o arquivo
  if [[ "$env_name" == tst ]]; then cp "$src/api.env" "$1/api.env"; return; fi
  need_env PRD_DOMAIN
  local domain="$PRD_DOMAIN" mail="${MAIL_FROM:-nao-responda@$PRD_DOMAIN}"
  [[ "$domain" =~ ^[a-z0-9]([a-z0-9.-]*[a-z0-9])?\.[a-z]{2,}$ ]] || die "PRD_DOMAIN invalido (esperado: app.exemplo.com.br, sem https://): $domain" 2
  [[ "$mail" =~ ^[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Za-z]{2,}$ ]] || die "MAIL_FROM invalido: $mail" 2
  sed -e "s|{{DOMAIN}}|$domain|g" -e "s|{{MAIL_FROM}}|$mail|g" "$src/api.env.tmpl" >"$1/api.env"
  ! grep -q '{{' "$1/api.env" || die "api.env.tmpl tem marcador sem valor" 2
}

write_secrets() { # write_secrets ARQUIVO
  local providers=0
  {
    printf 'DATABASE_URL=%s\n' "$(strip "$DATABASE_URL")"
    if group_ok OIDC_ISSUER OIDC_CLIENT_ID OIDC_CLIENT_SECRET; then emit OIDC_ISSUER OIDC_CLIENT_ID OIDC_CLIENT_SECRET; providers=1; fi
    if group_ok GOOGLE_CLIENT_ID GOOGLE_CLIENT_SECRET; then emit GOOGLE_CLIENT_ID GOOGLE_CLIENT_SECRET; providers=1; fi
    if group_ok GITHUB_CLIENT_ID GITHUB_CLIENT_SECRET; then emit GITHUB_CLIENT_ID GITHUB_CLIENT_SECRET; providers=1; fi
    if [[ -n "${SESSION_KEY:-}" ]]; then emit SESSION_KEY; fi
    if [[ -n "${BREVO_API_KEY:-}" ]]; then printf 'BREVO_API_KEY=%s\n' "$(printf '%s' "$BREVO_API_KEY" | tr -d '[:space:]')"; fi
  } >"$1"
  if [[ "$providers" -eq 1 && -z "${SESSION_KEY:-}" ]]; then die "com login social, defina tambem SESSION_KEY (openssl rand -base64 32)" 2; fi
}

umask 077
work="$(mktemp -d)"
trap 'rm -rf "$work"' EXIT
members=(compose.yaml Caddyfile api.env secrets.env deploy.sh)

cp "$src/compose.yaml" "$src/Caddyfile" "$work/"
cp "$here/deploy.sh" "$work/deploy.sh"
render_api_env "$work"
write_secrets "$work/secrets.env"
if [[ "$env_name" == prd ]]; then # a producao leva os scripts de backup (o deploy.sh faz backup antes de migrar)
  mkdir -p "$work/lib"
  cp "$here/backup.sh" "$here/restore.sh" "$here/restore-drill.sh" "$work/"
  cp "$here/lib/common.sh" "$here/lib/retention.sh" "$work/lib/"
  members+=(backup.sh restore.sh restore-drill.sh lib/common.sh lib/retention.sh)
fi
tar -C "$work" -cf "$out" "${members[@]}"
log "pacote $env_name: ${members[*]}"
