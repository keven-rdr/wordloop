#!/usr/bin/env bash
# Endurece a VM de producao. Idempotente: rodar de novo nao muda nada. Com --check so relata o que diverge (sai 1).
#
#   sudo bash harden-vm.sh --admin-user ubuntu            aplica
#   sudo bash harden-vm.sh --admin-user ubuntu --check    so confere (use no cron ou depois de mexer na VM)
#
# O que faz: SSH so por chave (sem senha, sem root) liberado so para o admin e o usuario de deploy; firewall (ufw)
# so com 22, 80 e 443; fail2ban no SSH; atualizacoes de seguranca automaticas; sysctl de rede; rotacao de log do
# Docker; cron do backup diario. TRAVA ANTI-LOCKOUT: recusa se o admin nao tem chave autorizada, e testa o sshd
# (sshd -t) antes de recarregar; se o teste falhar, desfaz.
# Teste: HARDEN_ROOT=/pasta grava tudo sob essa pasta e dispensa o root.
set -euo pipefail

here="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
. "$here/lib/common.sh"

ROOT="${HARDEN_ROOT:-}"
APP_DIR="${HARDEN_APP_DIR:-/srv/wordloop/prd}"
DEPLOY_USER="${HARDEN_DEPLOY_USER:-deploy}"
CHECK=0
ADMIN=""
DRIFT=0
LAST_CHANGED=0

while [[ $# -gt 0 ]]; do
  case "$1" in
    --check) CHECK=1 ;;
    --admin-user) ADMIN="${2:-}"; shift ;;
    *) die "uso: harden-vm.sh --admin-user <usuario> [--check]" 2 ;;
  esac
  shift
done
[[ -n "$ADMIN" ]] || die "uso: harden-vm.sh --admin-user <usuario> [--check]" 2
[[ "$ADMIN" =~ ^[a-z_][a-z0-9_-]{0,31}$ ]] || die "nome de usuario invalido: $ADMIN" 2
[[ -n "$ROOT" || "$(id -u)" -eq 0 ]] || die "rode como root (sudo)" 2

drift() { # drift MENSAGEM: registra uma divergencia
  DRIFT=$((DRIFT + 1))
  log "diverge: $1"
}

# want_file CAMINHO MODO: o conteudo vem da entrada padrao. Grava so se for diferente (e so fora do --check).
# LAST_CHANGED fica 1 quando gravou de fato.
want_file() {
  local path="$ROOT$1" mode="$2" new
  new="$(cat)"
  LAST_CHANGED=0
  if [[ -f "$path" && "$(cat "$path")" == "$new" ]]; then return 0; fi
  drift "$1"
  [[ "$CHECK" -eq 0 ]] || return 0
  LAST_CHANGED=1
  mkdir -p "$(dirname "$path")"
  printf '%s\n' "$new" >"$path"
  chmod "$mode" "$path"
  log "atualizado: $1"
}

has_key() { # o usuario tem pelo menos uma chave autorizada?
  local f="$ROOT/home/$1/.ssh/authorized_keys"
  [[ -s "$f" ]] && grep -qE '^(ssh-|ecdsa-|sk-|command=)' "$f"
}

harden_ssh() {
  has_key "$ADMIN" || die "o usuario $ADMIN nao tem chave SSH autorizada em /home/$ADMIN/.ssh/authorized_keys: aplicar travaria voce fora da VM" 2
  has_key "$DEPLOY_USER" || log "aviso: $DEPLOY_USER ainda sem chave (rode o bootstrap-vm.sh); o deploy so passa a funcionar depois"
  local f=/etc/ssh/sshd_config.d/99-wordloop.conf before=""
  [[ ! -f "$ROOT$f" ]] || before="$(cat "$ROOT$f")"
  want_file "$f" 644 <<EOF
# Gerenciado por deploy/scripts/harden-vm.sh
PasswordAuthentication no
KbdInteractiveAuthentication no
PermitRootLogin no
PubkeyAuthentication yes
X11Forwarding no
MaxAuthTries 3
LoginGraceTime 30
AllowUsers $ADMIN $DEPLOY_USER
EOF
  [[ "$LAST_CHANGED" -eq 1 ]] || return 0
  if ! sshd -t; then
    if [[ -n "$before" ]]; then printf '%s\n' "$before" >"$ROOT$f"; else rm -f "$ROOT$f"; fi
    die "sshd -t reprovou a configuracao: desfiz a mudanca de SSH (nada foi recarregado)" 4
  fi
  systemctl reload ssh
  log "sshd recarregado"
}

harden_packages() {
  local p missing=()
  for p in ufw fail2ban unattended-upgrades; do
    dpkg -s "$p" >/dev/null 2>&1 || missing+=("$p")
  done
  [[ "${#missing[@]}" -gt 0 ]] || return 0
  drift "pacotes ausentes: ${missing[*]}"
  [[ "$CHECK" -eq 1 ]] || DEBIAN_FRONTEND=noninteractive apt-get install -y -qq "${missing[@]}"
}

harden_firewall() {
  local status want rule
  status="$(ufw status 2>/dev/null || true)"
  want=(22/tcp 80/tcp 443/tcp 443/udp)
  local ok=1
  grep -q '^Status: active' <<<"$status" || ok=0
  for rule in "${want[@]}"; do
    grep -qE "^${rule}[[:space:]]+ALLOW" <<<"$status" || ok=0
  done
  [[ "$ok" -eq 0 ]] || return 0
  drift "firewall (ufw) fora do esperado"
  if [[ "$CHECK" -eq 1 ]]; then return 0; fi
  ufw default deny incoming >/dev/null
  ufw default allow outgoing >/dev/null
  for rule in "${want[@]}"; do ufw allow "$rule" >/dev/null; done # 22 liberado ANTES de ligar: nao se tranca para fora
  ufw --force enable >/dev/null
  log "firewall ligado: so 22, 80 e 443"
}

harden_files() {
  want_file /etc/fail2ban/jail.d/wordloop.local 644 <<'EOF'
[sshd]
enabled = true
maxretry = 5
findtime = 10m
bantime = 1h
EOF
  want_file /etc/apt/apt.conf.d/20auto-upgrades 644 <<'EOF'
APT::Periodic::Update-Package-Lists "1";
APT::Periodic::Unattended-Upgrade "1";
EOF
  want_file /etc/sysctl.d/98-wordloop-hardening.conf 644 <<'EOF'
net.ipv4.conf.all.rp_filter = 1
net.ipv4.conf.default.rp_filter = 1
net.ipv4.conf.all.accept_redirects = 0
net.ipv6.conf.all.accept_redirects = 0
net.ipv4.tcp_syncookies = 1
kernel.kptr_restrict = 2
EOF
  [[ "$LAST_CHANGED" -eq 0 ]] || sysctl -q --system || true
  want_file /etc/docker/daemon.json 644 <<'EOF'
{
  "log-driver": "json-file",
  "log-opts": { "max-size": "10m", "max-file": "3" }
}
EOF
  [[ "$LAST_CHANGED" -eq 0 ]] || log "reinicie o Docker numa janela tranquila para valer a rotacao de log"
}

harden_backup_cron() {
  want_file /etc/cron.d/wordloop-backup 644 <<EOF
# Backup diario do banco (criptografado, para o R2). Gerenciado por harden-vm.sh.
MAILTO=""
15 3 * * * $DEPLOY_USER $APP_DIR/backup.sh >>$APP_DIR/backup.log 2>&1
EOF
  want_file /etc/logrotate.d/wordloop-backup 644 <<EOF
$APP_DIR/backup.log {
  weekly
  rotate 8
  compress
  missingok
  notifempty
}
EOF
}

harden_ssh
harden_packages
harden_firewall
harden_files
harden_backup_cron

if [[ "$DRIFT" -eq 0 ]]; then
  log "VM conforme: nada a fazer"
elif [[ "$CHECK" -eq 1 ]]; then
  log "$DRIFT divergencia(s): rode sem --check para aplicar"
  exit 1
else
  log "VM endurecida ($DRIFT ajuste(s) aplicado(s))"
fi
