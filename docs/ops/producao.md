# Produção (beta): guia do dono

Tudo aqui foi entregue como **código e testes**; **nada foi provisionado**. Este guia é a sequência que *você* executa quando tiver as contas. Nunca cole chaves, senhas ou connection strings no chat, no Git ou em issue: elas vivem nos *secrets* do GitHub (Environment `prd`) e na VM.

## O que existe

| Peça | Onde | Para quê |
|---|---|---|
| Compose, Caddy, modelos de env | `deploy/prd/` | o mesmo desenho de tst: Caddy + API + web; banco **externo** (`DATABASE_URL`); documentação interativa da API **desligada** |
| Deploy de prd | `.github/workflows/deploy-prd.yml` | manual, com **aprovação** (Environment `prd`), só de commit que está em `develop`/`main`, smoke test no fim |
| Pacote de deploy | `deploy/scripts/build-package.sh` | monta compose, Caddyfile, `api.env` preenchido, `secrets.env` e scripts de backup |
| Deploy na VM | `deploy/scripts/deploy.sh` | **backup antes de migrar**; se o backup falhar, o deploy para e o banco não é tocado |
| Backup | `deploy/scripts/backup.sh` | `pg_dump` → `age` → R2, nunca em texto claro no disco; confere o tamanho no destino; retenção 14 dias + 8 semanas |
| Restauração | `deploy/scripts/restore.sh`, `restore-drill.sh` | restaura só em banco **descartável**; o ensaio sobe um Postgres em Docker, restaura e apaga |
| Endurecimento da VM | `deploy/scripts/harden-vm.sh` | SSH só por chave, `ufw` (22/80/443), `fail2ban`, atualizações automáticas, cron do backup; `--check` só confere |
| Testes dos scripts | `deploy/scripts/tests/` (`bash deploy/scripts/tests/run.sh`) e `ops-ci.yml` | tudo com comandos de mentira: sem rede, sem Docker, sem banco |

## Decisões que são suas

- **Banco de produção:** o app só precisa de uma `DATABASE_URL` PostgreSQL. Pode ser o Neon (como em tst, mas um *projeto separado* do de tst) ou um Postgres próprio. Use sempre bancos **diferentes** em tst e prd: o `restore.sh` recusa restaurar no mesmo servidor e banco do `DATABASE_URL`.
- **Domínio de produção** (`PRD_DOMAIN`, ex.: `wordloop.com.br`) e **remetente** de e-mail (`MAIL_FROM`).
- **Contato do responsável** e **prazo de guarda dos backups** nos textos de privacidade (estão como rascunho, revisão jurídica pendente).

## Sequência de colocação no ar

1. **VM.** Crie a VM Ubuntu e o DNS (`A` para o IP). Gere a chave de deploy **no seu computador**: `ssh-keygen -t ed25519 -f wl-prd-deploy -C wl-prd-deploy` (a privada vai para o secret `DEPLOY_SSH_KEY`).
2. **Bootstrap:** `WORDLOOP_ENV=prd sudo bash bootstrap-vm.sh "<chave pública>"` (swap, Docker, usuário `deploy` com chave de comando forçado, pasta `/srv/wordloop/prd`).
3. **Endurecimento:** com a sua chave de administrador já autorizada na VM, rode `sudo bash harden-vm.sh --admin-user <seu-usuário> --check` (só relata) e depois sem `--check`. A trava recusa aplicar se o admin não tem chave (você não se tranca para fora), e testa o `sshd` antes de recarregar. **Abra um segundo terminal SSH antes de aplicar** e só feche o primeiro depois de conferir que o segundo entra.
4. **Versões na VM:** em `/srv/wordloop/prd/versions.env` (modo 600), copie `deploy/prd/versions.env.example` com o seu `PRD_DOMAIN` e tags `sha-...` de um commit que já passou pelo tst.
5. **Secrets e variáveis** (*Settings → Environments → prd*): secrets `DEPLOY_SSH_KEY`, `DEPLOY_HOST`, `DEPLOY_KNOWN_HOSTS` (saída de `ssh-keyscan`), `DATABASE_URL`; opcionais conforme `deploy/prd/secrets.env.example`; variáveis `PRD_DOMAIN` e, se quiser, `MAIL_FROM`. Em **Required reviewers**, coloque você: é isso que impõe a aprovação.
6. **Backup** (seção abaixo) antes do primeiro deploy.
7. **Deploy:** *Actions → deploy-prd → Run workflow* com `api_tag` e `web_tag` (`sha-` + 40 caracteres). Aprove no Environment. O smoke test confere o commit servido, a página, e que a documentação interativa **não** está exposta.
8. **Primeiro admin:** como em tst (`create-admin`, ver [ci-cd](../ci-cd.md)), em `/srv/wordloop/prd`, por SSH com a *sua* chave.

**Rollback:** *deploy-prd* com `action: rollback` volta as imagens à versão anterior. Migração **não** se desfaz (expand/contract): se uma migração precisar ser revertida, a saída é restaurar o backup (seção abaixo), com decisão sua.

## Backup e restauração

**Chaves (uma vez, no seu computador, nunca na VM):** `age-keygen -o wl-backup.key`. A linha `# public key: age1...` vai para o `backup.env` da VM; **o arquivo `wl-backup.key` fica com você**, em dois lugares seguros (gerenciador de senhas e um pendrive, por exemplo). **Sem essa chave privada o backup não abre.**

**R2:** um bucket privado (ex.: `wordloop-backups`) e uma chave de API restrita a ele. Na VM, copie `deploy/prd/backup.env.example` para `/srv/wordloop/prd/backup.env` (modo 600) e preencha. Opcional e recomendado: um check no healthchecks.io (`BACKUP_PING_URL`), que avisa quando o backup **deixa de rodar**.

**Agendar:** `harden-vm.sh` instala o cron (todo dia, 03:15 UTC, usuário `deploy`, log em `/srv/wordloop/prd/backup.log`). Teste uma vez à mão: `sudo -u deploy /srv/wordloop/prd/backup.sh`.

**Ensaio de restauração (mensal, e depois de mexer no backup):** num computador com Docker e `age`, `rclone` e `psql`/`pg_restore` instalados, com a chave privada e as variáveis `RCLONE_CONFIG_R2_*` do `backup.env` exportadas no terminal:

```
BACKUP_REMOTE=r2:wordloop-backups/prd BACKUP_AGE_IDENTITY_FILE=./wl-backup.key bash deploy/scripts/restore-drill.sh latest
```

Termina com `RESTORE_OK file=... migration=... tables=... users=... reviews=...`. O banco descartável é apagado ao final. Um backup que nunca foi restaurado não é backup.

**Perda real do banco:** crie o banco novo, aponte `RESTORE_DATABASE_URL` para ele e rode `restore.sh --disposable latest` (a flag confirma que o destino é o banco certo para ser sobrescrito); depois troque o `DATABASE_URL` no secret e faça o deploy.

**Retenção:** todos dos últimos 14 dias e o primeiro de cada uma das últimas 8 semanas; ajustável por `BACKUP_KEEP_DAILY`/`BACKUP_KEEP_WEEKLY`. A retenção só roda depois de o backup novo ser conferido no destino e nunca com menos de 3 backups na lista. **Exclusão de conta (LGPD) e backups:** o backup guarda a conta até sair da janela de retenção; os textos de privacidade devem dizer isso (prazo = a retenção escolhida).

## Verificação de rotina

- `sudo bash harden-vm.sh --admin-user <usuário> --check`: confere se a VM continua como deve (sai 1 se algo divergiu).
- `ops-ci` roda os testes dos scripts a cada mudança em `deploy/`.
- Uma vez por mês: ensaio de restauração e conferir que o último backup no R2 é de hoje.

## O que **não** foi feito (de propósito)

Provisionar VM, DNS, banco, bucket e chaves; configurar o Environment `prd` e seus revisores; revisão jurídica dos textos; observabilidade (Prometheus/Grafana). Cada um depende de conta ou decisão sua.
