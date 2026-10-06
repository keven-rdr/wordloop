# ADR 0018 — release-please por componente; build uma vez e promoção por re-etiqueta

**Status:** aceita (2026-10-06)

## Contexto
O que foi testado em tst deve ser o que roda em prd; sem commit de bot em laço.

## Decisão
release-please em monorepo (`api-vX`, `web-vX`); imagem `sha-<commit>` construída na `develop`; a mesma imagem é re-etiquetada com a versão; tst usa `git describe`.

## Consequências
Tags criadas pelo `GITHUB_TOKEN` não disparam workflows: a promoção fica no mesmo arquivo; atribuição por caminho a validar.

## Alternativas descartadas
Bump de patch com commit de bot (Contratos); semantic-release; rebuild por ambiente.

## Fonte
`docs/fase-4/04-engenharia.md`
