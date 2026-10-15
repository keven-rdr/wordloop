#!/usr/bin/env bash
# Comando FORCADO da chave SSH de deploy (ver bootstrap-vm.sh). Chamado pelo workflow deploy.yml:
#   ssh deploy@VM "<api_tag> <web_tag>" < pacote.tar    ('-' mantem a tag atual)   |   ssh deploy@VM "rollback"
# O pacote (tar na entrada padrao) traz a configuracao do repositorio e o secrets.env montado pelo CI:
# compose.yaml, Caddyfile, api.env, secrets.env e deploy.sh. Sem pacote, usa o que ja esta na VM.
# So aceita tags no formato seguro; nada do que chega pelo SSH vira comando.
# Ao chamar a mao, redirecione a entrada: ssh ... "rollback" < /dev/null
set -euo pipefail

DIR="${WORDLOOP_DIR:-/srv/wordloop/tst}"
read -r API_TAG WEB_TAG <<<"${SSH_ORIGINAL_COMMAND:-}"
cd "$DIR"

if [[ ! -t 0 ]]; then
  tmp="$(mktemp -d)"
  trap 'rm -rf "$tmp"' EXIT
  cat >"$tmp/pacote.tar"
  if [[ -s "$tmp/pacote.tar" ]]; then
    tar -xf "$tmp/pacote.tar" -C "$tmp" compose.yaml Caddyfile api.env secrets.env deploy.sh
    install -m 644 "$tmp/compose.yaml" "$tmp/api.env" "$DIR/"
    # Caddyfile e montado como ARQUIVO no contêiner: gravar no lugar (mesmo inode) para o contêiner enxergar a troca.
    cat "$tmp/Caddyfile" >"$DIR/Caddyfile"
    install -m 600 "$tmp/secrets.env" "$DIR/secrets.env"
    install -m 755 "$tmp/deploy.sh" "$DIR/deploy.sh.new" # vira o deploy.sh no fim: nao troca o script em execucao
  fi
fi

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
if [[ "${API_TAG:-}" != "rollback" && "${API_TAG:-}" != "-" ]]; then
  # Migracoes ANTES de subir a API nova (expand/contract). Nunca no rollback: migracao nao se desfaz.
  # So quando a API muda. O timeout existe porque uma imagem sem o subcomando `migrate` subiria o servidor e ficaria
  # presa; 15 min, porque uma migracao grande e lenta ate o Neon (cada instrucao e uma ida e volta de ~0,15 s):
  # com 120 s, o seed v0 (849 instrucoes) estourou por 8 s mesmo funcionando.
  timeout 900 docker compose --env-file versions.env run --rm api migrate
fi
compose up -d --remove-orphans
# Recarrega a configuracao do Caddy sem derrubar conexoes (valida antes de aplicar); aviso, nao erro, se ele ainda subia.
compose exec -T caddy caddy reload --config /etc/caddy/Caddyfile || echo "aviso: nao foi possivel recarregar o Caddy" >&2
docker image prune -f >/dev/null # disco pequeno: apaga imagens antigas sem uso
compose ps
[[ -f deploy.sh.new ]] && mv -f deploy.sh.new deploy.sh
exit 0
