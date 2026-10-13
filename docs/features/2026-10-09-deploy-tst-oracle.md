# Deploy de tst na VM Oracle Micro

> Arquivo: `docs/features/2026-10-09-deploy-tst-oracle.md` · Issue: #__ · Branch: `feat/deploy-tst-oracle` · Estado: BUILD

## Contexto
O primeiro ambiente publicado (tst) roda numa VM Oracle de 1 GB. Ver [ADR 0022](../adr/0022-vm-oracle-micro-neon-sem-keycloak-em-tst.md).

## Definições confirmadas
- Banco: Neon gerenciado. Keycloak: fora da VM por enquanto. Domínio próprio. IP público reservado `wordloop-tst`.
- Deploy **automático** em todo push na `develop` (pacote de configuração + `DATABASE_URL` do secret enviados por SSH; migração antes de subir a API); manual e rollback pelo `workflow_dispatch`.

## Decisões
- Sem `ufw`: o Docker publica portas por fora dele; o firewall é a Security List da Oracle (22 só do dono, 80 e 443 abertos).
- Sem `environment.json` no web por enquanto (o web ainda não o lê).
- Healthcheck da API: externo (`/api/v1/health/ready`); a imagem é distroless.

## Checklist de progresso
- [x] `deploy/tst/` (compose, Caddyfile, exemplos), `deploy/scripts/` (deploy, bootstrap), `deploy.yml`
- [x] VM (Ubuntu 20.04, sem capacidade para recriar): swap, Docker, usuário `deploy`
- [x] DNS `tst.wordloop.com.br` → `146.235.63.55`; Security List 80/443
- [x] Secrets `DEPLOY_SSH_KEY`, `DEPLOY_HOST`, `DEPLOY_KNOWN_HOSTS` no GitHub
- [ ] Primeiro deploy e rollback testados
- [ ] `npm run check` verde

## Verificação
`curl -I https://tst.<dominio>/` → 200; `curl https://tst.<dominio>/api/v1/version` → `env: tst`; `docker stats` < 500 MB; deploy com tag nova e rollback.
