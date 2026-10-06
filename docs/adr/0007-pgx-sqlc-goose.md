# ADR 0007 — pgx + sqlc + goose

**Status:** aceita (2026-10-06)

## Contexto
Quero SQL visível, tipado e migrações simples.

## Decisão
`pgx` v5, consultas em `.sql` compiladas por `sqlc`, migrações com `goose`; transação por Unit of Work no contexto.

## Consequências
Consultas dinâmicas pedem cuidado manual; sem ORM.

## Alternativas descartadas
GORM (reflexão, comportamento implícito); ent.

## Fonte
`docs/fase-3/03-arquitetura-features.md`
