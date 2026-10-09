#!/usr/bin/env bash
# Comando FORCADO da chave SSH de deploy (ver bootstrap-vm.sh). Chamado pelo workflow deploy.yml:
#   ssh deploy@VM "<api_tag> <web_tag>"   ('-' mantem a tag atual)   |   ssh deploy@VM "rollback"
# So aceita tags no formato seguro; nada do que chega pelo SSH vira comando.
set -euo pipefail

DIR="${WORDLOOP_DIR:-/srv/wordloop/tst}"
read -r API_TAG WEB_TAG <<<"${SSH_ORIGINAL_COMMAND:-}"
cd "$DIR"

compose() { docker compose --env-file versions.env "$@"; }

if [[ "${API_TAG:-}" == "rollback" ]]; then
  [[ -f versions.env.prev ]] || { echo "sem versao anterior para voltar" >&2; exit 2; }
  cp versions.env.prev versions.env
else
  tag_ok() { [[ "${1:-}" == "-" || "${1:-}" =~ ^[A-Za-z0-9._-]{1,128}$ ]]; }
  tag_ok "${API_TAG:-}" && tag_ok "${WEB_TAG:-}" || { echo "tag invalida" >&2; exit 2; }
  cp versions.env versions.env.prev # base do rollback
  [[ "$API_TAG" == "-" ]] || sed -i "s|^API_TAG=.*|API_TAG=$API_TAG|" versions.env
  [[ "$WEB_TAG" == "-" ]] || sed -i "s|^WEB_TAG=.*|WEB_TAG=$WEB_TAG|" versions.env
fi

compose pull
compose up -d --remove-orphans
docker image prune -f >/dev/null # disco pequeno: apaga imagens antigas sem uso
compose ps
