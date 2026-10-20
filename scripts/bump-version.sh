#!/usr/bin/env sh
# Sobe o patch da versao de um componente (api ou web) e grava o novo numero na develop.
# Uso: scripts/bump-version.sh api|web   (imprime so a nova versao, ex.: 1.0.7)
#
# Cada componente tem a sua contagem em <componente>/VERSION (MAJOR.MINOR.PATCH). A cada merge na develop que
# muda o componente, o patch sobe 1 e um commit "chore(release): ... [skip ci]" grava o numero. O commit sai de um
# worktree separado: a pasta de trabalho do CI (de onde a imagem e construida) nao muda. Se outro merge gravar
# antes, tenta de novo em cima do mais recente. Para subir minor ou major, edite o arquivo VERSION a mao.
set -eu

comp=${1:?uso: bump-version.sh api|web}
case "$comp" in api | web) ;; *) echo "componente invalido: $comp" >&2; exit 2 ;; esac
branch=${BUMP_BRANCH:-develop}
tmp=$(mktemp -d)
trap 'git worktree remove --force "$tmp/w" >/dev/null 2>&1 || true; rm -rf "$tmp"' EXIT

attempt=1
while [ "$attempt" -le 5 ]; do
	git fetch -q origin "$branch"
	git worktree add -q --detach "$tmp/w" "origin/$branch"
	file="$tmp/w/$comp/VERSION"
	cur=$(tr -d ' \r\n' <"$file")
	case "$cur" in
	[0-9]*.[0-9]*.[0-9]*) ;;
	*) echo "versao invalida em $comp/VERSION: '$cur'" >&2; exit 1 ;;
	esac
	next=$(echo "$cur" | awk -F. '{printf "%d.%d.%d", $1, $2, $3 + 1}')
	printf '%s\n' "$next" >"$file"
	git -C "$tmp/w" add "$comp/VERSION"
	git -C "$tmp/w" -c user.name="github-actions[bot]" -c user.email="41898282+github-actions[bot]@users.noreply.github.com" \
		commit -q -m "chore(release): atualizar versao de $comp para $next [skip ci]"
	if git -C "$tmp/w" push -q origin "HEAD:$branch" >&2; then
		echo "$next"
		exit 0
	fi
	git worktree remove --force "$tmp/w"
	attempt=$((attempt + 1))
	sleep "$attempt"
done
echo "nao consegui gravar a versao de $comp na $branch (5 tentativas)" >&2
exit 1
