# ADR 0017 — Monorepo público; `develop` e `main`; merge por squash

**Status:** aceita (2026-10-06)

## Contexto
Contrato, cliente gerado e deploy mudam juntos; Sonar Free e Actions favorecem repositório público.

## Decisão
`keven-rdr/wordloop`: `api/`, `web/`, `deploy/`, `docs/`, `content/`; PRs para `develop` com título em Conventional Commits; `main` aponta para o que está em prd e é avançada manualmente por fast-forward.

## Consequências
Conteúdo precisa de licença compatível (ADR 0024); `main` depende de disciplina manual.

## Alternativas descartadas
Dois repositórios; terceiro de infraestrutura; trunk-based puro.

## Fonte
`docs/fase-4/04-engenharia.md`
