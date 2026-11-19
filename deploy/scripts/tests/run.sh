#!/usr/bin/env bash
# Roda todos os testes dos scripts de operacao (test_*.sh). Sai com erro se qualquer um falhar.
#   bash deploy/scripts/tests/run.sh
set -uo pipefail

here="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
failed=0
for t in "$here"/test_*.sh; do
  printf '== %s\n' "$(basename "$t")"
  bash "$t" || failed=$((failed + 1))
done
if [[ "$failed" -ne 0 ]]; then
  printf 'scripts de operacao: %d arquivo(s) de teste com falha\n' "$failed" >&2
  exit 1
fi
echo "scripts de operacao: tudo verde"
