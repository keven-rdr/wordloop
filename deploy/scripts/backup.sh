#!/usr/bin/env bash
# Backup diario do banco: pg_dump -> age (criptografia) -> R2 (rclone). O dump NUNCA existe sem criptografia:
# o pg_dump alimenta o age direto por um pipe, e so o arquivo .age toca o disco.
#
#   backup.sh                 (a VM chama todo dia pelo cron; ver docs/ops/producao.md)
#
# Configuracao, por variavel de ambiente ou pelo arquivo backup.env ao lado do script (formato CHAVE=valor):
#   DATABASE_URL           connection string (se faltar, le do secrets.env ao lado do script)
#   BACKUP_AGE_RECIPIENT   CHAVE PUBLICA do age (age1...). A chave privada fica FORA da VM, com o dono.
#   BACKUP_REMOTE          destino do rclone, ex.: r2:wordloop-backups/prd  (o remote r2 vem de RCLONE_CONFIG_R2_*)
#   BACKUP_PREFIX          prefixo dos arquivos (padrao: wordloop-prd)
#   BACKUP_KEEP_DAILY      quantos dias guarda todos os backups (padrao 14)
#   BACKUP_KEEP_WEEKLY     quantas semanas guarda o primeiro de cada semana (padrao 8)
#   BACKUP_PING_URL        (opcional) healthchecks.io: avisa inicio, sucesso e falha
# Sai com codigo diferente de zero em qualquer falha, e nunca imprime senha.
set -euo pipefail

here="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
. "$here/lib/common.sh"
. "$here/lib/retention.sh"

for v in DATABASE_URL BACKUP_AGE_RECIPIENT BACKUP_REMOTE BACKUP_PREFIX BACKUP_KEEP_DAILY BACKUP_KEEP_WEEKLY BACKUP_PING_URL   RCLONE_CONFIG_R2_TYPE RCLONE_CONFIG_R2_PROVIDER RCLONE_CONFIG_R2_ACCESS_KEY_ID RCLONE_CONFIG_R2_SECRET_ACCESS_KEY RCLONE_CONFIG_R2_ENDPOINT; do
  set_from_file "$v" "$here/backup.env"
done
set_from_file DATABASE_URL "${BACKUP_SECRETS_FILE:-$here/secrets.env}"

PG_DUMP="${BACKUP_PG_DUMP:-pg_dump}"
AGE="${BACKUP_AGE:-age}"
RCLONE="${BACKUP_RCLONE:-rclone}"
CURL="${BACKUP_CURL:-curl}"
PREFIX="${BACKUP_PREFIX:-wordloop-prd}"
KEEP_DAILY="${BACKUP_KEEP_DAILY:-14}"
KEEP_WEEKLY="${BACKUP_KEEP_WEEKLY:-8}"
MIN_BYTES="${BACKUP_MIN_BYTES:-1024}"
NOW="${BACKUP_NOW:-$(date -u +%s)}"

notify() { # notify [/start|/fail]: avisa o healthchecks.io, se configurado; nunca derruba o backup
  [[ -n "${BACKUP_PING_URL:-}" ]] || return 0
  "$CURL" -fsS -m 10 --retry 2 -o /dev/null "${BACKUP_PING_URL}$1" || log "aviso: nao consegui avisar o healthchecks"
}

tmp=""
reported=0
cleanup() { [[ -z "$tmp" ]] || rm -rf "$tmp"; }
report_failure() { # avisa UMA vez, qualquer que seja o caminho da falha
  [[ "$reported" -eq 1 ]] || notify /fail
  reported=1
}
die() { # sobrescreve o die comum: toda falha avisa o healthchecks
  log "erro: $1"
  report_failure
  exit "${2:-1}"
}
failed() { # trap de erro inesperado (um comando que falhou sem mensagem propria)
  local code=$?
  trap - ERR
  log "backup FALHOU (codigo $code)"
  report_failure
  cleanup
  exit "$code"
}
trap failed ERR
trap cleanup EXIT

need_env DATABASE_URL BACKUP_AGE_RECIPIENT BACKUP_REMOTE
[[ "$BACKUP_AGE_RECIPIENT" == age1* ]] || die "BACKUP_AGE_RECIPIENT deve ser uma chave PUBLICA do age (comeca com age1)" 2
[[ "$KEEP_DAILY" =~ ^[0-9]+$ && "$KEEP_WEEKLY" =~ ^[0-9]+$ ]] || die "BACKUP_KEEP_DAILY e BACKUP_KEEP_WEEKLY devem ser numeros" 2
need_cmd "$PG_DUMP" "$AGE" "$RCLONE"

notify /start
umask 077
tmp="$(mktemp -d)"
stamp="$(date -u -d "@$NOW" +%Y%m%dT%H%M%SZ)"
file="$PREFIX-$stamp.dump.age"
log "banco: $(db_identity "$DATABASE_URL" | redact)"

# 1) dump criptografado: nada de arquivo intermediario em texto claro
"$PG_DUMP" --format=custom --no-owner --no-privileges --dbname="$DATABASE_URL" | "$AGE" --recipient "$BACKUP_AGE_RECIPIENT" --output "$tmp/$file"
bytes="$(wc -c <"$tmp/$file" | tr -d ' ')"
[[ "$bytes" -ge "$MIN_BYTES" ]] || die "o backup saiu com $bytes bytes (minimo $MIN_BYTES): algo deu errado no dump"

# 2) soma de verificacao ao lado do arquivo
(cd "$tmp" && sha256sum "$file" >"$file.sha256")

# 3) envia e CONFERE o tamanho no destino (copiar sem conferir nao e backup)
"$RCLONE" copyto "$tmp/$file" "$BACKUP_REMOTE/$file"
"$RCLONE" copyto "$tmp/$file.sha256" "$BACKUP_REMOTE/$file.sha256"
remote_bytes="$("$RCLONE" lsf --format s --files-only "$BACKUP_REMOTE/$file" | tr -d ' \r\n')"
[[ "$remote_bytes" == "$bytes" ]] || die "tamanho no destino ($remote_bytes) diferente do local ($bytes)"
log "enviado: $file ($bytes bytes)"

# 4) retencao: so depois de o novo estar conferido no destino, e nunca com menos de 3 backups na lista
mapfile -t existing < <("$RCLONE" lsf --files-only "$BACKUP_REMOTE" | grep -E '\.dump\.age$' || true)
if [[ "${#existing[@]}" -ge 3 ]]; then
  deleted=0
  while IFS= read -r old; do
    [[ -n "$old" && "$old" != "$file" ]] || continue
    "$RCLONE" deletefile "$BACKUP_REMOTE/$old"
    "$RCLONE" deletefile "$BACKUP_REMOTE/$old.sha256" || true
    deleted=$((deleted + 1))
  done < <(printf '%s\n' "${existing[@]}" | retention_delete "$KEEP_DAILY" "$KEEP_WEEKLY" "$NOW")
  log "retencao: $deleted antigo(s) apagado(s), ${#existing[@]} na lista"
fi

notify ""
log "backup ok: $file"
