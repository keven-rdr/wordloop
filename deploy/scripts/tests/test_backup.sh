#!/usr/bin/env bash
# Testes do backup.sh: tudo com comandos de mentira (sem banco, sem rede, sem Docker).
set -euo pipefail
here="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
SUITE="backup"
. "$here/lib.sh"

backup() { run_script bash "$here/../backup.sh"; }
FILE="wordloop-prd-20261118T030000Z.dump.age"

test_case "caminho feliz: envia o arquivo criptografado e a soma de verificacao"
backup_env
export BACKUP_NOW="$(epoch '2026-11-18 03:00:00')"
backup
assert_status "$CODE" 0 "backup deveria passar: $OUT"
assert_file "$SB/remote/prd/$FILE" "arquivo .age no destino"
assert_file "$SB/remote/prd/$FILE.sha256" "soma de verificacao no destino"
(cd "$SB/remote/prd" && sha256sum --check --status "$FILE.sha256") && ok || bad "a soma enviada nao confere com o arquivo"
assert_contains "$(calls)" "age --recipient age1testrecipient" "criptografou para a chave publica"
assert_contains "$(calls)" "--dbname=postgres://app:s3cr3t@db.prod.example/wordloop?sslmode=require" "dump do banco certo"
assert_not_contains "$(cat "$SB/remote/prd/$FILE")" "PLAINTEXT-DUMP" "o texto claro nunca chega ao destino"
assert_not_contains "$OUT" "s3cr3t" "a senha nao vai para o log"
assert_contains "$OUT" "backup ok" "mensagem final"

test_case "o dump nunca vira arquivo em texto claro"
backup_env
export BACKUP_NOW="$(epoch '2026-11-18 03:00:00')"
export TMPDIR="$SB/work"
backup
assert_status "$CODE" 0 "$OUT"
leftover="$(find "$SB/work" -type f | wc -l | tr -d ' ')"
assert_eq "$leftover" "0" "a pasta temporaria e apagada ao final"
unset TMPDIR

test_case "falha no pg_dump: sai com erro, nada e enviado e o healthchecks e avisado uma vez"
backup_env
export BACKUP_NOW="$(epoch '2026-11-18 03:00:00')" STUB_DUMP_FAIL=1 BACKUP_PING_URL=https://hc.test/abc
backup
assert_nonzero "$CODE" "deveria falhar"
assert_eq "$(ls -1 "$SB/remote/prd" | wc -l | tr -d ' ')" "0" "nada no destino"
assert_eq "$(count_calls 'hc.test/abc/fail')" "1" "um unico aviso de falha"
assert_contains "$(calls)" "hc.test/abc/start" "avisou o inicio"

test_case "dump pequeno demais e recusado"
backup_env
export BACKUP_NOW="$(epoch '2026-11-18 03:00:00')" STUB_DUMP_BYTES=3
backup
assert_nonzero "$CODE" "deveria falhar"
assert_contains "$OUT" "minimo" "explica o motivo"
assert_eq "$(count_calls 'rclone copyto')" "0" "nao enviou nada"

test_case "tamanho diferente no destino derruba o backup e impede a retencao"
backup_env
export BACKUP_NOW="$(epoch '2026-11-18 03:00:00')" STUB_SIZE_LIE=1
for d in 20261101 20261102 20261103; do : >"$SB/remote/prd/wordloop-prd-${d}T030000Z.dump.age"; done
backup
assert_nonzero "$CODE" "deveria falhar"
assert_contains "$OUT" "diferente do local"
assert_eq "$(count_calls 'rclone deletefile')" "0" "nao apagou backup antigo"

test_case "chave que nao e publica do age e recusada antes de tocar no banco"
backup_env
export BACKUP_AGE_RECIPIENT="AGE-SECRET-KEY-1XYZ"
backup
assert_status "$CODE" 2
assert_contains "$OUT" "PUBLICA"
assert_eq "$(count_calls 'pg_dump')" "0" "nem chamou o pg_dump"

test_case "configuracao faltando: diz qual variavel"
backup_env
unset BACKUP_REMOTE
backup
assert_status "$CODE" 2
assert_contains "$OUT" "BACKUP_REMOTE"

test_case "DATABASE_URL vem do secrets.env sem executar o arquivo (URL com &)"
backup_env
unset DATABASE_URL
export BACKUP_NOW="$(epoch '2026-11-18 03:00:00')"
printf 'DATABASE_URL=postgres://u:pw@h.example/db?sslmode=require&channel_binding=require\n' >"$SB/secrets.env"
export BACKUP_SECRETS_FILE="$SB/secrets.env"
backup
assert_status "$CODE" 0 "$OUT"
assert_contains "$(calls)" "--dbname=postgres://u:pw@h.example/db?sslmode=require&channel_binding=require"

test_case "retencao: apaga o que passou da politica e leva a soma junto"
backup_env
export BACKUP_NOW="$(epoch '2026-11-18 03:00:00')" BACKUP_KEEP_DAILY=3 BACKUP_KEEP_WEEKLY=4
for d in 20261110 20261109 20260901; do
  : >"$SB/remote/prd/wordloop-prd-${d}T030000Z.dump.age"
  : >"$SB/remote/prd/wordloop-prd-${d}T030000Z.dump.age.sha256"
done
backup
assert_status "$CODE" 0 "$OUT"
assert_no_file "$SB/remote/prd/wordloop-prd-20261110T030000Z.dump.age" "terca da mesma semana"
assert_no_file "$SB/remote/prd/wordloop-prd-20261110T030000Z.dump.age.sha256" "a soma vai junto"
assert_no_file "$SB/remote/prd/wordloop-prd-20260901T030000Z.dump.age" "fora das semanas guardadas"
assert_file "$SB/remote/prd/wordloop-prd-20261109T030000Z.dump.age" "primeiro da semana fica"
assert_file "$SB/remote/prd/$FILE" "o novo fica"

test_case "retencao: com menos de 3 backups na lista, nao apaga nada"
backup_env
export BACKUP_NOW="$(epoch '2026-11-18 03:00:00')" BACKUP_KEEP_DAILY=1 BACKUP_KEEP_WEEKLY=1
: >"$SB/remote/prd/wordloop-prd-20200101T030000Z.dump.age"
backup
assert_status "$CODE" 0 "$OUT"
assert_file "$SB/remote/prd/wordloop-prd-20200101T030000Z.dump.age" "so 2 na lista: nada apagado"
assert_eq "$(count_calls 'rclone deletefile')" "0"

test_case "retencao: outros prefixos e arquivos alheios no destino nao sao tocados"
backup_env
export BACKUP_NOW="$(epoch '2026-11-18 03:00:00')" BACKUP_KEEP_DAILY=1 BACKUP_KEEP_WEEKLY=1
for n in outro-20200101T030000Z.dump.age README.txt; do : >"$SB/remote/prd/$n"; done
: >"$SB/remote/prd/wordloop-prd-20200102T030000Z.dump.age"
backup
assert_status "$CODE" 0 "$OUT"
assert_file "$SB/remote/prd/README.txt" "arquivo alheio"
assert_no_file "$SB/remote/prd/wordloop-prd-20200102T030000Z.dump.age" "o nosso antigo vai embora"

test_case "healthchecks: avisa inicio e fim, e falha ao avisar nao derruba o backup"
backup_env
export BACKUP_NOW="$(epoch '2026-11-18 03:00:00')" BACKUP_PING_URL=https://hc.test/abc STUB_CURL_FAIL=1
backup
assert_status "$CODE" 0 "um ping que falha nao pode matar o backup: $OUT"
assert_contains "$OUT" "aviso" "registra o aviso"
assert_eq "$(count_calls 'hc.test/abc/fail')" "0" "nao e falha de backup"

summary
