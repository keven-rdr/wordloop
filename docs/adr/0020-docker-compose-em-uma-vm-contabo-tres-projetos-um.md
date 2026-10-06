# ADR 0020 — Docker Compose em uma VM Contabo: três projetos, um Postgres, Caddy

**Status:** aceita (2026-10-06)

## Contexto
Orçamento mínimo; ambientes tst e prd na mesma VM.

## Decisão
Projetos `shared` (Caddy, PostgreSQL com 3 bancos, Keycloak), `tst` e `prd`; só o Caddy publica portas; configuração em runtime (`environment.json`); imagens por tag+digest; backup `pg_dump`+`age` para o Cloudflare R2.

## Consequências
VM única é ponto único de falha; tst e prd compartilham o servidor de banco.

## Alternativas descartadas
Kubernetes leve; um Postgres por ambiente.

## Fonte
`docs/fase-4/deploy/deploy-skeleton.md`
