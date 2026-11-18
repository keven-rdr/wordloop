#!/usr/bin/env bash
# Testes do build-package.sh: o pacote que o CI manda para a VM (configuracao + secrets.env montado dos secrets).
set -euo pipefail
here="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
SUITE="pacote"
. "$here/lib.sh"

SECRET_VARS=(DATABASE_URL OIDC_ISSUER OIDC_CLIENT_ID OIDC_CLIENT_SECRET GOOGLE_CLIENT_ID GOOGLE_CLIENT_SECRET
  GITHUB_CLIENT_ID GITHUB_CLIENT_SECRET SESSION_KEY BREVO_API_KEY PRD_DOMAIN MAIL_FROM)

pkg_setup() { # pkg_setup NOME: sandbox com variaveis limpas e a base de DATABASE_URL
  test_case "$1"
  unset "${SECRET_VARS[@]}"
  export DATABASE_URL='postgres://u:pw@h.example/db?sslmode=require&channel_binding=require'
  export PRD_DOMAIN=app.exemplo.com.br
}

build() { run_script bash "$here/../build-package.sh" "$1" "$SB/p.tar"; }
member() { tar -xOf "$SB/p.tar" "$1"; }
members() { tar -tf "$SB/p.tar" | sort | tr '\n' ' '; }

pkg_setup "prd: leva a configuracao, os scripts de backup e o api.env preenchido"
build prd
assert_status "$CODE" 0 "$OUT"
assert_eq "$(members)" "Caddyfile api.env backup.sh compose.yaml deploy.sh lib/common.sh lib/retention.sh restore-drill.sh restore.sh secrets.env " "membros"
assert_contains "$(member api.env)" "OIDC_REDIRECT_URL=https://app.exemplo.com.br/api/v1/auth/callback"
assert_contains "$(member api.env)" "WEBAUTHN_RP_ID=app.exemplo.com.br"
assert_contains "$(member api.env)" "MAIL_FROM=nao-responda@app.exemplo.com.br" "remetente padrao"
assert_not_contains "$(member api.env)" "{{" "nenhum marcador sobrando"
assert_contains "$(member compose.yaml)" "APP_ENV: prd"
assert_eq "$(member secrets.env)" "DATABASE_URL=postgres://u:pw@h.example/db?sslmode=require&channel_binding=require" "so o banco"
assert_not_contains "$OUT" "pw@" "segredo fora do log"

pkg_setup "prd: MAIL_FROM do ambiente vence o padrao"
export MAIL_FROM=ola@wordloop.com.br
build prd
assert_status "$CODE" 0 "$OUT"
assert_contains "$(member api.env)" "MAIL_FROM=ola@wordloop.com.br"

pkg_setup "prd: sem dominio ou com dominio estranho nao gera pacote"
unset PRD_DOMAIN
build prd
assert_status "$CODE" 2
assert_contains "$OUT" "PRD_DOMAIN"
for bad in 'https://x.com' 'x.com/a' 'a|b.com' 'semponto' 'x.com; rm -rf /'; do
  export PRD_DOMAIN="$bad"
  build prd
  assert_status "$CODE" 2 "deveria recusar: $bad"
done
export PRD_DOMAIN=x.com MAIL_FROM='a|b@x.com'
build prd
assert_status "$CODE" 2 "MAIL_FROM invalido"

pkg_setup "tst: leva so a configuracao de tst, sem scripts de backup"
build tst
assert_status "$CODE" 0 "$OUT"
assert_eq "$(members)" "Caddyfile api.env compose.yaml deploy.sh secrets.env " "membros de tst"
assert_contains "$(member api.env)" "tst.wordloop.com.br"

pkg_setup "login: grupo completo entra; pela metade e erro; com provedor exige SESSION_KEY"
export GOOGLE_CLIENT_ID=gid GOOGLE_CLIENT_SECRET=gsecret SESSION_KEY=chave== BREVO_API_KEY=" br evo
"
build prd
assert_status "$CODE" 0 "$OUT"
sec="$(member secrets.env)"
assert_contains "$sec" "GOOGLE_CLIENT_ID=gid"
assert_contains "$sec" "GOOGLE_CLIENT_SECRET=gsecret"
assert_contains "$sec" "SESSION_KEY=chave=="
assert_contains "$sec" "BREVO_API_KEY=brevo" "espacos e quebras saem da chave"
unset GOOGLE_CLIENT_SECRET
build prd
assert_status "$CODE" 2 "grupo pela metade"
assert_contains "$OUT" "GOOGLE_CLIENT_ID"
export GOOGLE_CLIENT_SECRET=gsecret
unset SESSION_KEY
build prd
assert_status "$CODE" 2 "provedor sem SESSION_KEY"
assert_contains "$OUT" "SESSION_KEY"

pkg_setup "sem DATABASE_URL nao gera pacote; ambiente invalido tambem nao"
unset DATABASE_URL
build prd
assert_status "$CODE" 2
assert_contains "$OUT" "DATABASE_URL"
export DATABASE_URL=x
build dev
assert_status "$CODE" 2

summary
