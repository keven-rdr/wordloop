#!/usr/bin/env bash
# Prepara a VM Ubuntu (Oracle Micro, 1 GB) UMA vez. Rode como root:  sudo bash bootstrap-vm.sh "<chave-publica-de-deploy>"
# Idempotente. A chave publica de deploy fica restrita ao comando deploy.sh (sem shell, sem tunel).
set -euo pipefail

PUBKEY="${1:?uso: bootstrap-vm.sh \"ssh-ed25519 AAAA... deploy\"}"
BASE=/srv/wordloop/tst

# 1) Swap de 2 GB: com 1 GB de RAM o Docker e o pull de imagens estouram sem isso.
if ! swapon --show | grep -q /swapfile; then
  fallocate -l 2G /swapfile
  chmod 600 /swapfile
  mkswap /swapfile >/dev/null
  swapon /swapfile
  grep -q '^/swapfile' /etc/fstab || echo '/swapfile none swap sw 0 0' >>/etc/fstab
fi
echo 'vm.swappiness=10' >/etc/sysctl.d/99-wordloop.conf
sysctl -q -p /etc/sysctl.d/99-wordloop.conf

# 2) Docker Engine + plugin compose pelo repositorio apt oficial (sem script baixado e executado).
if ! command -v docker >/dev/null; then
  apt-get update -qq
  apt-get install -y -qq ca-certificates curl gnupg
  install -m 0755 -d /etc/apt/keyrings
  curl -fsSL https://download.docker.com/linux/ubuntu/gpg -o /etc/apt/keyrings/docker.asc
  chmod a+r /etc/apt/keyrings/docker.asc
  . /etc/os-release
  echo "deb [arch=$(dpkg --print-architecture) signed-by=/etc/apt/keyrings/docker.asc] https://download.docker.com/linux/ubuntu ${VERSION_CODENAME} stable" \
    >/etc/apt/sources.list.d/docker.list
  apt-get update -qq
  apt-get install -y -qq docker-ce docker-ce-cli containerd.io docker-compose-plugin
fi

# 3) Usuario de deploy so com a chave forcada para o deploy.sh.
id deploy >/dev/null 2>&1 || useradd -m -s /bin/bash deploy
usermod -aG docker deploy
install -d -m 755 -o deploy -g deploy "$BASE"
install -d -m 700 -o deploy -g deploy /home/deploy/.ssh
echo "command=\"$BASE/deploy.sh\",no-port-forwarding,no-agent-forwarding,no-X11-forwarding,no-pty $PUBKEY" \
  >/home/deploy/.ssh/authorized_keys
chown deploy:deploy /home/deploy/.ssh/authorized_keys
chmod 600 /home/deploy/.ssh/authorized_keys

echo "ok: copie compose.yaml, Caddyfile, deploy.sh e crie versions.env e secrets.env em $BASE"
