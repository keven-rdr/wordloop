#!/usr/bin/env sh
# Versao do produto: base de VERSION (MAJOR.MINOR, editada a mao) + numero de merges na develop.
# Cada merge (squash) soma 1; o mesmo commit tem a mesma versao em dev, tst e prd. Precisa do historico completo.
set -eu
root=$(cd "$(dirname "$0")/.." && pwd)
base=$(tr -d ' \r\n' < "$root/VERSION")
echo "$base.$(git -C "$root" rev-list --count --first-parent HEAD)"
