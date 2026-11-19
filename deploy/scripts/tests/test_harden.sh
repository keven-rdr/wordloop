#!/usr/bin/env bash
# Testes do harden-vm.sh: a "VM" e uma pasta (HARDEN_ROOT); ufw, dpkg, apt-get, sshd, systemctl e sysctl sao de mentira.
set -euo pipefail
here="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
SUITE="harden"
. "$here/lib.sh"

harden_setup() { # harden_setup NOME: sandbox + VM de mentira com admin "ubuntu" ja com chave
  test_case "$1"
  export HARDEN_ROOT="$SB/vm"
  mkdir -p "$HARDEN_ROOT/home/ubuntu/.ssh" "$HARDEN_ROOT/home/deploy/.ssh"
  echo 'ssh-ed25519 AAAAC3 admin' >"$HARDEN_ROOT/home/ubuntu/.ssh/authorized_keys"
  echo 'command="/srv/wordloop/prd/deploy.sh",no-pty ssh-ed25519 AAAAC3 deploy' >"$HARDEN_ROOT/home/deploy/.ssh/authorized_keys"
  unset STUB_SSHD_FAIL
  stub dpkg <<'EOF'
[[ "$1" == "-s" ]] && grep -qx "$2" "$SB/installed" 2>/dev/null
EOF
  stub apt-get <<'EOF'
echo "apt-get $*" >>"$SB/calls.log"
for a in "$@"; do case "$a" in -*|install) ;; *) echo "$a" >>"$SB/installed" ;; esac; done
EOF
  stub ufw <<'EOF'
echo "ufw $*" >>"$SB/calls.log"
case "$1" in
  status)
    if [[ -f "$SB/ufw.active" ]]; then echo "Status: active"; else echo "Status: inactive"; fi
    [[ ! -f "$SB/ufw.rules" ]] || sed 's|$|   ALLOW       Anywhere|' "$SB/ufw.rules" ;;
  allow) echo "$2" >>"$SB/ufw.rules" ;;
  --force) touch "$SB/ufw.active" ;;
esac
EOF
  stub sshd <<'EOF'
echo "sshd $*" >>"$SB/calls.log"
[[ -z "${STUB_SSHD_FAIL:-}" ]]
EOF
  stub systemctl <<'EOF'
echo "systemctl $*" >>"$SB/calls.log"
EOF
  stub sysctl <<'EOF'
echo "sysctl $*" >>"$SB/calls.log"
EOF
}

harden() { run_script bash "$here/../harden-vm.sh" "$@"; }

harden_setup "aplica: SSH so por chave, firewall, fail2ban, atualizacoes e cron do backup"
harden --admin-user ubuntu
assert_status "$CODE" 0 "$OUT"
conf="$(cat "$HARDEN_ROOT/etc/ssh/sshd_config.d/99-wordloop.conf")"
assert_contains "$conf" "PasswordAuthentication no"
assert_contains "$conf" "PermitRootLogin no"
assert_contains "$conf" "AllowUsers ubuntu deploy"
assert_contains "$(calls)" "sshd -t" "testou o sshd antes de recarregar"
assert_contains "$(calls)" "systemctl reload ssh"
assert_contains "$(calls)" "ufw allow 22/tcp"
assert_contains "$(calls)" "ufw allow 443/udp" "HTTP/3"
assert_contains "$(calls)" "apt-get install" "instalou o que faltava"
assert_file "$HARDEN_ROOT/etc/fail2ban/jail.d/wordloop.local"
assert_file "$HARDEN_ROOT/etc/apt/apt.conf.d/20auto-upgrades"
assert_contains "$(cat "$HARDEN_ROOT/etc/cron.d/wordloop-backup")" "15 3 * * * deploy /srv/wordloop/prd/backup.sh"
assert_contains "$(cat "$HARDEN_ROOT/etc/docker/daemon.json")" "max-size"
u22="$(grep -n 'ufw allow 22/tcp' "$SB/calls.log" | head -n1 | cut -d: -f1)"
en="$(grep -n 'ufw --force enable' "$SB/calls.log" | head -n1 | cut -d: -f1)"
[[ -n "$u22" && -n "$en" && "$u22" -lt "$en" ]] && ok || bad "a porta 22 precisa ser liberada ANTES de ligar o firewall"

harden_setup "idempotente: a segunda rodada nao muda nada e o --check passa"
harden --admin-user ubuntu
assert_status "$CODE" 0 "$OUT"
before="$(find "$HARDEN_ROOT/etc" -type f -exec sha256sum {} + | sort)"
: >"$SB/calls.log"
harden --admin-user ubuntu
assert_status "$CODE" 0 "$OUT"
assert_contains "$OUT" "conforme"
assert_eq "$(find "$HARDEN_ROOT/etc" -type f -exec sha256sum {} + | sort)" "$before" "arquivos intactos"
assert_eq "$(count_calls 'systemctl reload')" "0" "nao recarrega o sshd sem mudanca"
assert_eq "$(count_calls 'apt-get')" "0" "nao reinstala"
harden --admin-user ubuntu --check
assert_status "$CODE" 0 "$OUT"

harden_setup "--check numa VM crua so relata: sai 1 e nao altera nada"
harden --admin-user ubuntu --check
assert_status "$CODE" 1
assert_contains "$OUT" "diverge"
assert_no_file "$HARDEN_ROOT/etc" "nenhum arquivo criado"
assert_eq "$(count_calls 'apt-get')" "0"
assert_eq "$(count_calls 'ufw allow')" "0"
assert_eq "$(count_calls 'systemctl')" "0"

harden_setup "--check acusa o desvio depois que alguem reabre a senha no SSH"
harden --admin-user ubuntu
printf 'PasswordAuthentication yes\n' >"$HARDEN_ROOT/etc/ssh/sshd_config.d/99-wordloop.conf"
harden --admin-user ubuntu --check
assert_status "$CODE" 1
assert_contains "$OUT" "99-wordloop.conf"

harden_setup "trava anti-lockout: sem chave do admin, nao aplica nada"
: >"$HARDEN_ROOT/home/ubuntu/.ssh/authorized_keys"
harden --admin-user ubuntu
assert_status "$CODE" 2
assert_contains "$OUT" "travaria"
assert_no_file "$HARDEN_ROOT/etc" "nada gravado"
assert_eq "$(count_calls 'ufw')" "0" "firewall intacto"

harden_setup "sshd -t reprova: desfaz a mudanca e nao recarrega"
printf 'AllowUsers antigo\n' >/dev/null
mkdir -p "$HARDEN_ROOT/etc/ssh/sshd_config.d"
printf 'MaxAuthTries 6\n' >"$HARDEN_ROOT/etc/ssh/sshd_config.d/99-wordloop.conf"
export STUB_SSHD_FAIL=1
harden --admin-user ubuntu
assert_status "$CODE" 4
assert_contains "$OUT" "desfiz"
assert_eq "$(cat "$HARDEN_ROOT/etc/ssh/sshd_config.d/99-wordloop.conf")" "MaxAuthTries 6" "configuracao anterior restaurada"
assert_eq "$(count_calls 'systemctl reload')" "0"
assert_eq "$(count_calls 'ufw')" "0" "parou antes do firewall"

harden_setup "sshd -t reprova numa VM sem o arquivo: remove o que criou"
export STUB_SSHD_FAIL=1
harden --admin-user ubuntu
assert_status "$CODE" 4
assert_no_file "$HARDEN_ROOT/etc/ssh/sshd_config.d/99-wordloop.conf"

harden_setup "argumentos invalidos"
harden
assert_status "$CODE" 2
harden --admin-user 'x; rm -rf /'
assert_status "$CODE" 2
harden --admin-user ubuntu --nao-existe
assert_status "$CODE" 2

summary
