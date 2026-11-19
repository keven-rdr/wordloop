#!/usr/bin/env bash
# Apoio dos testes dos scripts de operacao. Sem rede, sem Docker, sem banco: cada teste cria um "sandbox" com comandos
# de mentira (pg_dump, age, rclone, psql...) no PATH, que gravam o que receberam num diario para o teste conferir.

PASS=0
FAIL=0
CURRENT=""

ok() { PASS=$((PASS + 1)); }
bad() {
  FAIL=$((FAIL + 1))
  printf '  FALHOU [%s]: %s\n' "$CURRENT" "$1" >&2
}

assert_eq() { if [[ "$1" == "$2" ]]; then ok; else bad "${3:-esperava '$2', veio '$1'}"; fi; }
assert_contains() { if [[ "$1" == *"$2"* ]]; then ok; else bad "${3:-esperava conter '$2' em: $1}"; fi; }
assert_not_contains() { if [[ "$1" != *"$2"* ]]; then ok; else bad "${3:-nao devia conter '$2'}"; fi; }
assert_file() { if [[ -f "$1" ]]; then ok; else bad "${2:-arquivo ausente: $1}"; fi; }
assert_no_file() { if [[ ! -e "$1" ]]; then ok; else bad "${2:-nao devia existir: $1}"; fi; }
assert_status() { if [[ "$1" -eq "$2" ]]; then ok; else bad "${3:-codigo $1, esperava $2}"; fi; }
assert_nonzero() { if [[ "$1" -ne 0 ]]; then ok; else bad "${2:-devia falhar}"; fi; }

# test_case NOME: marca o teste atual e prepara um sandbox novo em $SB.
test_case() {
  CURRENT="$1"
  SB="$(mktemp -d)"
  mkdir -p "$SB/bin" "$SB/remote" "$SB/work"
  : >"$SB/calls.log"
  export SB PATH="$SB/bin:$ORIGINAL_PATH"
  make_stubs
}

# Comandos de mentira. Cada um escreve uma linha em $SB/calls.log e se comporta como o de verdade no que importa.
make_stubs() {
  stub pg_dump <<'EOF'
echo "pg_dump $*" >>"$SB/calls.log"
[[ -z "${STUB_DUMP_FAIL:-}" ]] || { echo "pg_dump: erro simulado" >&2; exit 1; }
head -c "${STUB_DUMP_BYTES:-4096}" /dev/zero | tr '\0' 'D'
printf 'PLAINTEXT-DUMP'
EOF
  # age de mentira: "criptografa" embrulhando em base64 (o texto claro nunca aparece), e desfaz na decifragem.
  stub age <<'EOF'
echo "age $*" >>"$SB/calls.log"
mode=enc; out=""; in=""
while [[ $# -gt 0 ]]; do
  case "$1" in
    --decrypt|-d) mode=dec ;;
    --output|-o) out="$2"; shift ;;
    --recipient|-r|--identity|-i) shift ;;
    *) in="$1" ;;
  esac
  shift
done
if [[ "$mode" == enc ]]; then
  { printf 'AGE-ENCRYPTED:'; base64 -w0; } >"$out"
else
  [[ -z "${STUB_AGE_FAIL:-}" ]] || { echo "age: chave errada" >&2; exit 1; }
  tail -c +15 "$in" | base64 -d
fi
EOF
  stub rclone <<'EOF'
echo "rclone $*" >>"$SB/calls.log"
cmd="$1"; shift
dest_dir() { printf '%s' "$SB/remote"; }
rpath() { printf '%s/%s' "$(dest_dir)" "${1#*:*/}"; }
case "$cmd" in
  copyto)
    src="$1"; dst="$2"
    if [[ "$src" == r2:* ]]; then cp "$(rpath "$src")" "$dst"; else cp "$src" "$(rpath "$dst")"; fi
    if [[ -n "${STUB_CORRUPT:-}" && "$dst" != r2:* ]]; then printf 'X' >>"$dst"; fi ;;
  lsf)
    fmt=""; files=""
    while [[ $# -gt 0 ]]; do
      case "$1" in --format) fmt="$2"; shift ;; --files-only) files=1 ;; *) target="$1" ;; esac
      shift
    done
    p="$(rpath "$target")"
    if [[ -d "$p" ]]; then ls -1 "$p"
    elif [[ "$fmt" == s ]]; then
      size="$(wc -c <"$p" | tr -d ' ')"
      [[ -z "${STUB_SIZE_LIE:-}" ]] || size=$((size + 7))
      echo "$size"
    else basename "$p"; fi ;;
  deletefile) rm -f "$(rpath "$1")" ;;
esac
EOF
  stub curl <<'EOF'
echo "curl $*" >>"$SB/calls.log"
[[ -z "${STUB_CURL_FAIL:-}" ]] || exit 22
EOF
  stub pg_restore <<'EOF'
echo "pg_restore $*" >>"$SB/calls.log"
cat >"$SB/restored.dump"
[[ -z "${STUB_RESTORE_FAIL:-}" ]] || { echo "pg_restore: erro simulado" >&2; exit 1; }
EOF
  stub psql <<'EOF'
echo "psql $*" >>"$SB/calls.log"
case "$*" in
  *max\(version_id\)*) echo "${STUB_LATEST:-17}" ;;
  *goose_db_version*) echo "${STUB_MIGRATIONS:-17}" ;;
  *information_schema*) echo "${STUB_TABLES:-24}" ;;
  *app_user*) echo "${STUB_USERS:-3}" ;;
  *review_log*) echo "${STUB_REVIEWS:-120}" ;;
esac
EOF
}

# stub NOME: cria um executavel em $SB/bin com o corpo lido da entrada padrao.
stub() {
  { printf '#!/usr/bin/env bash\n'; cat; } >"$SB/bin/$1"
  chmod +x "$SB/bin/$1"
}

# run_script CMD...: roda e guarda saida e codigo em $OUT e $CODE (sem derrubar o teste por causa de `set -e`).
run_script() {
  set +e
  OUT="$("$@" 2>&1)"
  CODE=$?
  set -e
}

calls() { cat "$SB/calls.log"; }
count_calls() { grep -c "$1" "$SB/calls.log" || true; }

summary() {
  printf '%s: %d ok, %d falha(s)\n' "${SUITE:-testes}" "$PASS" "$FAIL"
  [[ "$FAIL" -eq 0 ]]
}

TESTS_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# backup_env: configuracao minima do backup.sh / restore.sh dentro do sandbox atual (zera os STUB_* do teste anterior).
backup_env() {
  unset BACKUP_PING_URL BACKUP_PREFIX BACKUP_KEEP_DAILY BACKUP_KEEP_WEEKLY BACKUP_MIN_BYTES BACKUP_NOW \
    STUB_DUMP_FAIL STUB_DUMP_BYTES STUB_AGE_FAIL STUB_CORRUPT STUB_SIZE_LIE STUB_CURL_FAIL STUB_RESTORE_FAIL \
    STUB_MIGRATIONS STUB_LATEST STUB_TABLES STUB_USERS STUB_REVIEWS
  mkdir -p "$SB/remote/prd"
  printf 'AGE-SECRET-KEY-FAKE\n' >"$SB/key.txt"
  export DATABASE_URL='postgres://app:s3cr3t@db.prod.example/wordloop?sslmode=require'
  export BACKUP_AGE_RECIPIENT=age1testrecipient BACKUP_REMOTE=r2:bucket/prd BACKUP_SECRETS_FILE="$SB/sem-secrets.env"
  export RESTORE_DATABASE_URL='postgres://postgres:pw@127.0.0.1:5432/drill?sslmode=disable' BACKUP_AGE_IDENTITY_FILE="$SB/key.txt"
}

epoch() { date -u -d "$1" +%s; }

# make_backup "2026-11-18 03:00:00": roda o backup.sh como se fosse naquele instante.
make_backup() { BACKUP_NOW="$(epoch "$1")" bash "$TESTS_DIR/../backup.sh" >/dev/null 2>&1; }

ORIGINAL_PATH="$PATH"
