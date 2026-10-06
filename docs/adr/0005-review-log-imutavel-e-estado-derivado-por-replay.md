# ADR 0005 — `review_log` imutável e estado derivado por replay

**Status:** aceita (2026-10-06)

## Contexto
Precisamos re-treinar/otimizar o algoritmo, comparar estratégias e auditar.

## Decisão
Tabela só-INSERT (gatilho bloqueia UPDATE/DELETE), UUID do cliente, estado antes/depois, sugerido × escolhido, scheduler+versão; correção por evento `review_void`.

## Consequências
Armazena mais; exclusão de conta (LGPD) exige rotina controlada e registrada.

## Alternativas descartadas
Só estado atual (impede replay); eventos em tabela mutável.

## Fonte
`docs/fase-1/01-srs.md; fase-3/schema-learning.sql`
