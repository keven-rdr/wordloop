#!/usr/bin/env bash
# Politica de retencao dos backups, como funcao pura (testavel sem rede nem relogio):
#   retention_delete KEEP_DAILY KEEP_WEEKLY AGORA_EM_SEGUNDOS < lista-de-nomes
# Os nomes sao PREFIXO-AAAAMMDDTHHMMSSZ.dump.age (um por linha). Guarda:
#   1) todo backup dos ultimos KEEP_DAILY dias;
#   2) o PRIMEIRO backup de cada semana ISO, nas ultimas KEEP_WEEKLY semanas.
# Imprime, um por linha, os nomes a APAGAR. Nome que nao segue o padrao nunca e apagado (nao e nosso).

# epoch_of NOME: segundos (UTC) do carimbo no nome; vazio se nao tem carimbo.
epoch_of() {
  local stamp
  stamp="$(printf '%s' "$1" | sed -nE 's/.*-([0-9]{8})T([0-9]{2})([0-9]{2})([0-9]{2})Z\.dump\.age$/\1 \2:\3:\4/p')"
  [[ -n "$stamp" ]] || return 0
  date -u -d "${stamp:0:4}-${stamp:4:2}-${stamp:6:2} ${stamp:9}" +%s 2>/dev/null || true
}

retention_delete() {
  local keep_daily="$1" keep_weekly="$2" now="$3" name epoch week
  local -a names=() epochs=()
  declare -A first_of_week=()
  while IFS= read -r name; do
    [[ -n "$name" ]] || continue
    epoch="$(epoch_of "$name")"
    [[ -n "$epoch" ]] || continue # nao e nosso: nunca apaga
    names+=("$name")
    epochs+=("$epoch")
  done
  # o mais antigo de cada semana ISO e o que a semana guarda
  local i
  for i in "${!names[@]}"; do
    week="$(date -u -d "@${epochs[$i]}" +%G-W%V)"
    if [[ -z "${first_of_week[$week]:-}" || "${epochs[$i]}" -lt "${first_of_week[$week]%%|*}" ]]; then
      first_of_week[$week]="${epochs[$i]}|${names[$i]}"
    fi
  done
  local weekly_cutoff=$((now - keep_weekly * 7 * 86400)) daily_cutoff=$((now - keep_daily * 86400))
  for i in "${!names[@]}"; do
    [[ "${epochs[$i]}" -ge "$daily_cutoff" ]] && continue
    week="$(date -u -d "@${epochs[$i]}" +%G-W%V)"
    if [[ "${first_of_week[$week]#*|}" == "${names[$i]}" && "${epochs[$i]}" -ge "$weekly_cutoff" ]]; then continue; fi
    printf '%s\n' "${names[$i]}"
  done
}
