# ADR 0002 — Pacote por funcionalidade e limites medidos no CI

**Status:** aceita (2026-10-06)

## Contexto
O Contratos tem pastas por tipo técnico (`services/` 53 arquivos, `queries/` 50, `schemas/` 58) e `router.tsx` com 1.278 linhas.

## Decisão
Pacote/pasta por funcionalidade; arquivo ≤ 300 linhas, função ≤ 40 (Go) / 60 (TS), complexidade ≤ 10, ≤ 12 arquivos por pacote Go; `depguard`, `go-arch-lint`, ESLint e `check-structure` reprovam no CI; hook `guard.mjs` bloqueia na edição.

## Consequências
Regras viram código executável, não só texto; custo inicial de configurar as ferramentas.

## Alternativas descartadas
Só guia escrito (já divergiu no Contratos); só revisão humana.

## Fonte
`docs/fase-3/03b-estrutura-pastas.md`
