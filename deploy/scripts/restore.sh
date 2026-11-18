#!/usr/bin/env bash
# Restaura um backup (.dump.age) num banco DESCARTAVEL e confere o resultado. Serve para o ensaio de restauracao
# e para o dia em que o banco de verdade se perder (ai o destino e o banco novo, escolhido com cuidado).
#
#   restore.sh [--disposable] <latest | nome-do-arquivo.dump.age>
#
# Variaveis (ou o arquivo backup.env ao lado):
#   RESTORE_DATABASE_URL        banco DE DESTINO (obrigatorio). Nunca o de producao.
#   BACKUP_AGE_IDENTITY_FILE    arquivo com a chave PRIVADA do age (fica com o dono, fora da VM)
#   BACKUP_REMOTE, BACKUP_PREFIX   iguais as do backup.sh
# Trava de seguranca: se RESTORE_DATABASE_URL aponta para o mesmo servidor e banco do DATABASE_URL, recusa.
set -euo pipefail

here="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
. "$here/lib/common.sh"

for v in BACKUP_REMOTE BACKUP_PREFIX RESTORE_DATABASE_URL BACKUP_AGE_IDENTITY_FILE RCLONE_CONFIG_R2_TYPE RCLONE_CONFIG_R2_PROVIDER \
  RCLONE_CONFIG_R2_ACCESS_KEY_ID RCLONE_CONFIG_R2_SECRET_ACCESS_KEY RCLONE_CONFIG_R2_ENDPOINT; do
  set_from_file "$v" "$here/backup.env"
done
set_from_file DATABASE_URL "${BACKUP_SECRETS_FILE:-$here/secrets.env}"

AGE="${BACKUP_AGE:-age}"
RCLONE="${BACKUP_RCLONE:-rclone}"
PG_RESTORE="${BACKUP_PG_RESTORE:-pg_restore}"
PSQL="${BACKUP_PSQL:-psql}"
PREFIX="${BACKUP_PREFIX:-wordloop-prd}"

disposable=0
if [[ "${1:-}" == "--disposable" ]]; then
  disposable=1
  shift
fi
target="${1:-}"
[[ -n "$target" ]] || die "uso: restore.sh [--disposable] <latest | arquivo.dump.age>" 2

need_env BACKUP_REMOTE RESTORE_DATABASE_URL BACKUP_AGE_IDENTITY_FILE
[[ -f "$BACKUP_AGE_IDENTITY_FILE" ]] || die "chave privada nao encontrada: $BACKUP_AGE_IDENTITY_FILE" 2
need_cmd "$AGE" "$RCLONE" "$PG_RESTORE" "$PSQL"

# Trava: nunca restaurar em cima do banco de producao.
if [[ -n "${DATABASE_URL:-}" && "$(db_identity "$DATABASE_URL")" == "$(db_identity "$RESTORE_DATABASE_URL")" ]]; then
  die "RESTORE_DATABASE_URL e o MESMO banco de DATABASE_URL: recusado (restaure num banco descartavel)" 3
fi
if [[ "$disposable" -ne 1 ]]; then
  die "confirme que o destino e descartavel: rode com --disposable (destino: $(db_identity "$RESTORE_DATABASE_URL"))" 3
fi

tmp="$(mktemp -d)"
trap 'rm -rf "$tmp"' EXIT
umask 077

if [[ "$target" == "latest" ]]; then
  target="$("$RCLONE" lsf --files-only "$BACKUP_REMOTE" | grep -E "^${PREFIX}-[0-9]{8}T[0-9]{6}Z\.dump\.age$" | sort | tail -n 1 || true)"
  [[ -n "$target" ]] || die "nenhum backup $PREFIX-*.dump.age em $BACKUP_REMOTE" 4
fi
[[ "$target" =~ ^[A-Za-z0-9._-]+\.dump\.age$ ]] || die "nome de arquivo invalido: $target" 2
log "restaurando $target em $(db_identity "$RESTORE_DATABASE_URL")"

"$RCLONE" copyto "$BACKUP_REMOTE/$target" "$tmp/$target"
"$RCLONE" copyto "$BACKUP_REMOTE/$target.sha256" "$tmp/$target.sha256"
(cd "$tmp" && sha256sum --check --status "$target.sha256") || die "soma de verificacao NAO confere: backup corrompido ou adulterado" 5

# --clean --if-exists: o ensaio pode rodar de novo no mesmo banco descartavel
"$AGE" --decrypt --identity "$BACKUP_AGE_IDENTITY_FILE" "$tmp/$target" |
  "$PG_RESTORE" --no-owner --no-privileges --clean --if-exists --exit-on-error --dbname="$RESTORE_DATABASE_URL"

# Conferencia: as migracoes chegaram e ha dados. O resumo vai para o log; nada de dado pessoal.
q() { "$PSQL" "$RESTORE_DATABASE_URL" --no-psqlrc --tuples-only --no-align --command "$1" | tr -d ' \r\n'; }
migrations="$(q 'SELECT count(*) FROM goose_db_version')"
latest="$(q 'SELECT max(version_id) FROM goose_db_version')"
tables="$(q "SELECT count(*) FROM information_schema.tables WHERE table_schema = 'public'")"
users="$(q 'SELECT count(*) FROM app_user')"
reviews="$(q 'SELECT count(*) FROM review_log')"
[[ "$migrations" -gt 0 ]] || die "o banco restaurado nao tem migracoes: o backup nao parece ser do wordloop" 6
log "restauracao ok: $tables tabelas, migracao $latest, $users pessoas, $reviews respostas"
printf 'RESTORE_OK file=%s migration=%s tables=%s users=%s reviews=%s\n' "$target" "$latest" "$tables" "$users" "$reviews"
