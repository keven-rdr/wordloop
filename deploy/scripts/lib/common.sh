#!/usr/bin/env bash
# Funcoes comuns dos scripts de operacao (backup, restauracao, endurecimento). Use com:  . "$here/lib/common.sh"
# Nada aqui imprime segredo: connection strings passam por `redact` antes de irem a qualquer log.

log() { printf '%s %s\n' "$(date -u +%Y-%m-%dT%H:%M:%SZ)" "$*" >&2; }

# die "mensagem" [codigo]: registra e sai (o codigo 1 e o padrao).
die() {
  log "erro: $1"
  exit "${2:-1}"
}

need_env() {
  local v
  for v in "$@"; do
    [[ -n "${!v:-}" ]] || die "variavel ausente ou vazia: $v" 2
  done
}

need_cmd() {
  local c
  for c in "$@"; do
    command -v "$c" >/dev/null 2>&1 || die "comando nao encontrado: $c" 2
  done
}

# Troca a senha de uma connection string por *** (postgres://usuario:***@host/banco).
redact() { sed -E 's#(postgres(ql)?://[^:/@ ]+:)[^@ ]+@#\1***@#g'; }

# load_var NOME ARQUIVO: le NOME=valor de um arquivo no formato do secrets.env SEM executa-lo (uma URL com `&` ou `$`
# quebraria um `source`). Devolve vazio se nao achar. Aceita valor entre aspas simples ou duplas.
load_var() {
  local name="$1" file="$2" line
  [[ -f "$file" ]] || return 0
  line="$(grep -m1 -E "^${name}=" "$file" || true)"
  [[ -n "$line" ]] || return 0
  line="${line#*=}"
  line="${line%$'\r'}"
  if [[ "$line" =~ ^\"(.*)\"$ || "$line" =~ ^\'(.*)\'$ ]]; then line="${BASH_REMATCH[1]}"; fi
  printf '%s' "$line"
}

# set_from_file NOME ARQUIVO: se NOME ainda nao esta no ambiente, traz do arquivo.
set_from_file() {
  local name="$1" file="$2" value
  [[ -n "${!name:-}" ]] && return 0
  value="$(load_var "$name" "$file")"
  [[ -z "$value" ]] || export "$name=$value"
}

# Hostname e banco de uma connection string, para comparar dois bancos sem imprimir a senha.
db_identity() {
  local url="$1"
  url="${url#*://}"
  url="${url#*@}"
  url="${url%%\?*}"
  printf '%s' "$url"
}
