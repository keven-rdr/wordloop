#!/usr/bin/env bash
# Testes do restore.sh: o backup e feito pelo backup.sh (de mentira) e restaurado num "banco descartavel" de mentira.
set -euo pipefail
here="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
SUITE="restore"
. "$here/lib.sh"

restore() { run_script bash "$here/../restore.sh" "$@"; }
OLD="wordloop-prd-20261117T030000Z.dump.age"
NEW="wordloop-prd-20261118T030000Z.dump.age"

seed() { # dois backups: o de ontem e o de hoje
  make_backup '2026-11-17 03:00:00'
  make_backup '2026-11-18 03:00:00'
}

test_case "latest restaura o backup mais novo e imprime o resumo"
backup_env
seed
restore --disposable latest
assert_status "$CODE" 0 "$OUT"
assert_contains "$OUT" "RESTORE_OK file=$NEW migration=17 tables=24 users=3 reviews=120"
assert_contains "$(calls)" "rclone copyto r2:bucket/prd/$NEW " "baixou o mais novo"
assert_not_contains "$(calls)" "r2:bucket/prd/$OLD " "e nao o antigo"
assert_contains "$(calls)" "pg_restore --no-owner --no-privileges --clean --if-exists --exit-on-error" "restauracao repetivel"
assert_contains "$(cat "$SB/restored.dump")" "PLAINTEXT-DUMP" "o que chegou ao pg_restore e o dump decifrado"

test_case "arquivo explicito restaura exatamente aquele"
backup_env
seed
restore --disposable "$OLD"
assert_status "$CODE" 0 "$OUT"
assert_contains "$OUT" "file=$OLD"

test_case "sem --disposable recusa e nao restaura"
backup_env
seed
restore latest
assert_status "$CODE" 3
assert_contains "$OUT" "--disposable"
assert_eq "$(count_calls 'pg_restore')" "0"

test_case "recusa restaurar em cima do mesmo banco de producao (mesmo com --disposable)"
backup_env
seed
export RESTORE_DATABASE_URL='postgres://outra:senha@db.prod.example/wordloop?sslmode=require'
restore --disposable latest
assert_status "$CODE" 3
assert_contains "$OUT" "MESMO banco"
assert_eq "$(count_calls 'pg_restore')" "0"
assert_not_contains "$OUT" "outra:senha" "senha fora do log"

test_case "servidor igual mas banco diferente e permitido"
backup_env
seed
export RESTORE_DATABASE_URL='postgres://outra:senha@db.prod.example/ensaio?sslmode=require'
restore --disposable latest
assert_status "$CODE" 0 "$OUT"

test_case "backup adulterado: a soma de verificacao barra antes de restaurar"
backup_env
seed
printf 'X' >>"$SB/remote/prd/$NEW"
restore --disposable latest
assert_status "$CODE" 5
assert_contains "$OUT" "NAO confere"
assert_eq "$(count_calls 'pg_restore')" "0"

test_case "sem nenhum backup: avisa"
backup_env
restore --disposable latest
assert_status "$CODE" 4

test_case "nome de arquivo com caminho ou extensao errada e recusado"
backup_env
restore --disposable ../fora.dump.age
assert_status "$CODE" 2
restore --disposable "backup.txt"
assert_status "$CODE" 2
restore --disposable
assert_status "$CODE" 2
assert_contains "$OUT" "uso:"

test_case "banco restaurado sem migracoes nao e do wordloop"
backup_env
seed
export STUB_MIGRATIONS=0
restore --disposable latest
assert_status "$CODE" 6
assert_not_contains "$OUT" "RESTORE_OK"

test_case "falha do pg_restore ou chave errada nunca termina em RESTORE_OK"
backup_env
seed
export STUB_RESTORE_FAIL=1
restore --disposable latest
assert_nonzero "$CODE" "pg_restore falhou"
assert_not_contains "$OUT" "RESTORE_OK"
unset STUB_RESTORE_FAIL
export STUB_AGE_FAIL=1
restore --disposable latest
assert_nonzero "$CODE" "chave errada"
assert_not_contains "$OUT" "RESTORE_OK"

test_case "chave privada ausente: diz onde procurou"
backup_env
export BACKUP_AGE_IDENTITY_FILE="$SB/nao-existe.txt"
restore --disposable latest
assert_status "$CODE" 2
assert_contains "$OUT" "chave privada"

summary
