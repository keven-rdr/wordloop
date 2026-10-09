# ADR 0022 — tst na VM Oracle Micro, Postgres no Neon e sem Keycloak por enquanto

**Status:** aceita (2026-10-09). Revisa o [ADR 0020](0020-docker-compose-em-uma-vm-contabo-tres-projetos-um.md) **apenas para tst**.

## Contexto
A máquina disponível é a Oracle VM.Standard.E2.1.Micro (x86, 1 OCPU, 1 GB de RAM, 47 GB de disco). Não comporta Postgres + Keycloak + API + web, que o ADR 0020 previa em uma VM de 8-12 GB.

## Decisão
- tst roda **Caddy + API + web** na VM (~350 MB), com swap de 2 GB e `mem_limit` por contêiner.
- O Postgres é **gerenciado (Neon, região sa-east-1)**; a `DATABASE_URL` fica só em `secrets.env` na VM.
- **Sem Keycloak em tst** até haver máquina maior; o login existe só no dev local (`deploy/dev`).
- Imagens construídas no CI e publicadas no GHCR (`sha-<commit>`, multi-arquitetura); a VM só faz `pull`. Deploy por `deploy.yml` (manual) → SSH com chave de **comando forçado** → `deploy/scripts/deploy.sh`. Rollback = voltar a tag anterior.
- Mesma origem para web e API atrás do Caddy (TLS automático), sem CORS.

## Consequências
Neon free (0,5 GB) hiberna por inatividade e tem limite de tamanho. tst não testa o login real. A VM é ponto único de falha. Quando houver máquina maior, tst e prd voltam ao desenho do ADR 0020.

## Alternativas descartadas
Postgres e Keycloak na VM de 1 GB (sem folga, mesmo com swap); Ampere A1 (a Oracle costuma negar capacidade em São Paulo); build na própria VM.

## Fonte
`docs/fase-4/deploy/deploy-skeleton.md`; estratégia de build fora da VM e banco externo inspirada em projeto legado do autor.
