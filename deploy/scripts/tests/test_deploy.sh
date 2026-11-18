#!/usr/bin/env bash
# Testes do deploy.sh: docker de mentira que so registra as chamadas; o diretorio de deploy e o sandbox.
set -euo pipefail
here="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
SUITE="deploy"
. "$here/lib.sh"

setup() { # setup: sandbox com um "diretorio de deploy" pronto e docker de mentira
  test_case "$1"
  D="$SB/srv"
  mkdir -p "$D"
  printf 'API_TAG=1.0.0\nWEB_TAG=1.0.0\n' >"$D/versions.env"
  cp "$here/../deploy.sh" "$D/deploy.sh"
  stub docker <<'EOF'
echo "docker $*" >>"$SB/calls.log"
[[ -z "${STUB_MIGRATE_FAIL:-}" || "$*" != *migrate* ]] || exit 1
EOF
  unset SSH_ORIGINAL_COMMAND
  export WORDLOOP_DIR="$D"
}

deploy() { # deploy "<comando ssh>" [arquivo-de-entrada]
  export SSH_ORIGINAL_COMMAND="$1"
  run_script bash "$D/deploy.sh" <"${2:-/dev/null}"
}

with_backup() { # backup.sh de mentira + backup.env: ativa o backup antes de migrar
  printf '#!/usr/bin/env bash\necho "backup.sh" >>"$SB/calls.log"\n[[ -z "${STUB_BACKUP_FAIL:-}" ]]\n' >"$D/backup.sh"
  chmod +x "$D/backup.sh"
  : >"$D/backup.env"
}

setup "tag invalida e recusada (nada vira comando)"
deploy 'x; rm -rf /'
assert_status "$CODE" 2
deploy '1.0.1 $(id)'
assert_status "$CODE" 2
assert_eq "$(count_calls docker)" "0" "docker nem foi chamado"

setup "rollback sem versao anterior falha; com versao anterior volta"
deploy rollback
assert_status "$CODE" 2
printf 'API_TAG=0.9.0\nWEB_TAG=0.9.0\n' >"$D/versions.env.prev"
deploy rollback
assert_status "$CODE" 0 "$OUT"
assert_contains "$(cat "$D/versions.env")" "API_TAG=0.9.0"
assert_eq "$(count_calls migrate)" "0" "rollback nunca migra"

setup "API nova migra antes de subir; a anterior fica para rollback"
deploy "1.1.0 1.1.0"
assert_status "$CODE" 0 "$OUT"
assert_contains "$(cat "$D/versions.env")" "API_TAG=1.1.0"
assert_contains "$(cat "$D/versions.env.prev")" "API_TAG=1.0.0"
migrate_line="$(grep -n 'migrate' "$SB/calls.log" | head -n1 | cut -d: -f1)"
up_line="$(grep -n 'up -d' "$SB/calls.log" | head -n1 | cut -d: -f1)"
[[ -n "$migrate_line" && -n "$up_line" && "$migrate_line" -lt "$up_line" ]] && ok || bad "migrate deveria vir antes do up"

setup "so o front muda: sem migracao e sem backup"
with_backup
deploy "- 1.1.0"
assert_status "$CODE" 0 "$OUT"
assert_eq "$(count_calls migrate)" "0"
assert_eq "$(count_calls backup.sh)" "0"

setup "com backup configurado: backup roda antes de migrar"
with_backup
deploy "1.1.0 1.1.0"
assert_status "$CODE" 0 "$OUT"
b="$(grep -n 'backup.sh' "$SB/calls.log" | head -n1 | cut -d: -f1)"
m="$(grep -n 'migrate' "$SB/calls.log" | head -n1 | cut -d: -f1)"
[[ -n "$b" && -n "$m" && "$b" -lt "$m" ]] && ok || bad "backup deveria vir antes de migrate"

setup "backup falhou: deploy cancelado, sem migrar nem subir, versao registrada volta"
with_backup
export STUB_BACKUP_FAIL=1
deploy "1.1.0 1.1.0"
assert_status "$CODE" 3
assert_eq "$(count_calls migrate)" "0" "nao migrou"
assert_eq "$(count_calls 'up -d')" "0" "nao subiu"
assert_contains "$(cat "$D/versions.env")" "API_TAG=1.0.0" "versions.env voltou"
unset STUB_BACKUP_FAIL

setup "migracao que falha impede o up"
export STUB_MIGRATE_FAIL=1
deploy "1.1.0 1.1.0"
assert_nonzero "$CODE" "migracao falhou"
assert_eq "$(count_calls 'up -d')" "0"
unset STUB_MIGRATE_FAIL

pacote() { # pacote ARQUIVO [extras...]: tar com a configuracao (e os extras) montado numa pasta temporaria
  local out="$1" src="$SB/pkg"; shift
  mkdir -p "$src/lib"
  for f in compose.yaml Caddyfile api.env secrets.env; do echo "$f novo" >"$src/$f"; done
  cp "$here/../deploy.sh" "$src/deploy.sh"
  for f in "$@"; do echo "echo $f" >"$src/$f"; done
  (cd "$src" && tar -cf "$out" compose.yaml Caddyfile api.env secrets.env deploy.sh "$@")
}

setup "pacote: instala a configuracao e os scripts de backup da lista fechada"
pacote "$SB/p.tar" backup.sh lib/common.sh
deploy "- -" "$SB/p.tar"
assert_status "$CODE" 0 "$OUT"
assert_contains "$(cat "$D/compose.yaml")" "compose.yaml novo"
assert_file "$D/backup.sh" "backup.sh instalado"
assert_file "$D/lib/common.sh" "lib instalada"
[[ "$(uname -o)" == Msys ]] || assert_eq "$(stat -c %a "$D/secrets.env")" "600" "segredos so do dono (o NTFS nao tem modo Unix)"

setup "pacote: o que esta fora da lista nao e extraido"
pacote "$SB/p.tar" evil.sh
deploy "- -" "$SB/p.tar"
assert_status "$CODE" 0 "$OUT"
assert_no_file "$D/evil.sh" "membro fora da lista"

summary
