#!/usr/bin/env bash
# Testes de lib/common.sh e lib/retention.sh.
set -euo pipefail
here="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
SUITE="common e retencao"
. "$here/lib.sh"
. "$here/../lib/common.sh"
. "$here/../lib/retention.sh"

test_case "load_var le o valor sem executar o arquivo"
cat >"$SB/secrets.env" <<'EOF'
DATABASE_URL=postgres://u:p%40ss@host.neon.tech/neondb?sslmode=require&channel_binding=require
SESSION_KEY="abc=="
OUTRA='com aspas'
COM_CRLF=valor
EOF
printf 'COM_CRLF=valor\r\n' >>"$SB/secrets.env"
assert_eq "$(load_var DATABASE_URL "$SB/secrets.env")" "postgres://u:p%40ss@host.neon.tech/neondb?sslmode=require&channel_binding=require" "URL com & e %"
assert_eq "$(load_var SESSION_KEY "$SB/secrets.env")" "abc==" "aspas duplas saem"
assert_eq "$(load_var OUTRA "$SB/secrets.env")" "com aspas" "aspas simples saem"
assert_eq "$(load_var COM_CRLF "$SB/secrets.env")" "valor" "CRLF do Windows nao vai junto"
assert_eq "$(load_var NAO_EXISTE "$SB/secrets.env")" "" "chave ausente"
assert_eq "$(load_var X "$SB/nao-existe.env")" "" "arquivo ausente"

test_case "set_from_file nao sobrescreve o que ja esta no ambiente"
printf 'A=do-arquivo\nB=do-arquivo\n' >"$SB/e.env"
export A=do-ambiente
unset B
set_from_file A "$SB/e.env"
set_from_file B "$SB/e.env"
assert_eq "$A" "do-ambiente" "ambiente vence"
assert_eq "$B" "do-arquivo" "arquivo preenche o que falta"

test_case "redact esconde a senha da connection string"
out="$(printf 'banco: postgres://usuario:segredo123@host.neon.tech/db?sslmode=require\n' | redact)"
assert_contains "$out" "postgres://usuario:***@host.neon.tech" "senha trocada"
assert_not_contains "$out" "segredo123" "senha nao aparece"
out="$(printf 'postgresql://u:p@h/d\n' | redact)"
assert_contains "$out" "postgresql://u:***@h/d" "esquema postgresql"

test_case "db_identity compara servidor e banco sem a senha"
assert_eq "$(db_identity 'postgres://u:senha@host.neon.tech/neondb?sslmode=require')" "host.neon.tech/neondb"
assert_eq "$(db_identity 'postgres://postgres:x@127.0.0.1:5432/drill?sslmode=disable')" "127.0.0.1:5432/drill"

test_case "need_env e need_cmd falham com o nome do que faltou"
run_script bash -c ". '$here/../lib/common.sh'; need_env VAR_QUE_NAO_EXISTE"
assert_status "$CODE" 2
assert_contains "$OUT" "VAR_QUE_NAO_EXISTE"
run_script bash -c ". '$here/../lib/common.sh'; need_cmd comando-que-nao-existe-xyz"
assert_status "$CODE" 2
assert_contains "$OUT" "comando-que-nao-existe-xyz"

NOW="$(date -u -d '2026-11-18 12:00:00' +%s)"
names() { printf '%s\n' "$@"; }
test_case "retencao: guarda os dias recentes, o 1o de cada semana e apaga o resto"
deleted="$(names p-20261118T030000Z.dump.age p-20261117T030000Z.dump.age p-20261110T030000Z.dump.age \
  p-20261109T030000Z.dump.age p-20261108T030000Z.dump.age p-20260901T030000Z.dump.age | retention_delete 3 4 "$NOW")"
assert_contains "$deleted" "p-20261110T030000Z.dump.age" "terca da semana 46 nao e a primeira"
assert_contains "$deleted" "p-20260901T030000Z.dump.age" "mais velho que as semanas guardadas"
assert_not_contains "$deleted" "p-20261118T030000Z" "o de hoje fica"
assert_not_contains "$deleted" "p-20261117T030000Z" "o de ontem fica"
assert_not_contains "$deleted" "p-20261109T030000Z" "segunda, primeiro da semana 46, fica"
assert_not_contains "$deleted" "p-20261108T030000Z" "domingo, primeiro da semana 45, fica"

test_case "retencao: nome que nao e nosso nunca e apagado"
deleted="$(names README.txt p-20200101T030000Z.dump.age p-20200101T030000Z.dump.age.sha256 foto.jpg | retention_delete 1 1 "$NOW")"
assert_eq "$deleted" "p-20200101T030000Z.dump.age" "so o backup antigo vai; o resto nao tem o padrao"

test_case "retencao: lista vazia e uma so linha em branco nao quebram"
assert_eq "$(printf '' | retention_delete 14 8 "$NOW")" ""
assert_eq "$(printf '\n\n' | retention_delete 14 8 "$NOW")" ""

test_case "retencao: com tudo dentro da janela diaria, nada e apagado"
deleted="$(names p-20261118T030000Z.dump.age p-20261117T030000Z.dump.age p-20261116T030000Z.dump.age | retention_delete 14 8 "$NOW")"
assert_eq "$deleted" "" "14 dias de folga"

test_case "retencao: sem semanas guardadas, so valem os dias"
deleted="$(names p-20261118T030000Z.dump.age p-20261101T030000Z.dump.age | retention_delete 3 0 "$NOW")"
assert_eq "$deleted" "p-20261101T030000Z.dump.age"

summary
