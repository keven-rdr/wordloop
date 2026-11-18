#!/usr/bin/env bash
# Ensaio de restauracao: sobe um PostgreSQL descartavel em Docker, restaura o backup mais recente nele, confere e
# apaga tudo. E o criterio "backup restaurado em banco descartavel" do roadmap: um backup que nunca foi restaurado
# nao e backup. Rode uma vez por mes (e depois de mudar o backup).
#
#   restore-drill.sh [latest | arquivo.dump.age]        (precisa de Docker e da chave privada do age)
#
# Variaveis: as do restore.sh (BACKUP_REMOTE, BACKUP_AGE_IDENTITY_FILE...) e DRILL_PG_IMAGE (padrao postgres:17-alpine).
set -euo pipefail

here="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
. "$here/lib/common.sh"

DOCKER="${DRILL_DOCKER:-docker}"
IMAGE="${DRILL_PG_IMAGE:-postgres:17-alpine}"
TARGET="${1:-latest}"
need_cmd "$DOCKER"

name="wl-restore-drill-$$"
password="$(head -c 18 /dev/urandom | base64 | tr -dc 'A-Za-z0-9')"
cleanup() { "$DOCKER" rm -f "$name" >/dev/null 2>&1 || true; }
trap cleanup EXIT

log "subindo um PostgreSQL descartavel ($IMAGE)"
"$DOCKER" run -d --name "$name" -e POSTGRES_PASSWORD="$password" -e POSTGRES_DB=drill -p 127.0.0.1::5432 "$IMAGE" >/dev/null

port=""
for _ in $(seq 1 30); do
  if "$DOCKER" exec "$name" pg_isready -U postgres -d drill >/dev/null 2>&1; then
    port="$("$DOCKER" port "$name" 5432/tcp | head -n 1 | sed -E 's/.*:([0-9]+)$/\1/')"
    [[ -n "$port" ]] && break
  fi
  sleep 2
done
[[ -n "$port" ]] || die "o PostgreSQL descartavel nao ficou pronto" 7

export RESTORE_DATABASE_URL="postgres://postgres:${password}@127.0.0.1:${port}/drill?sslmode=disable"
bash "$here/restore.sh" --disposable "$TARGET"
log "ensaio concluido: o backup restaura. O banco descartavel foi apagado."
