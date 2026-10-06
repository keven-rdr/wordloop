# ADR 0006 — OpenAPI 3.0 primeiro; servidor `oapi-codegen`, cliente `hey-api`

**Status:** aceita (2026-10-06)

## Contexto
Contrato único sem desvio entre servidor e cliente (no Contratos tipos e serviços são escritos à mão).

## Decisão
`api/openapi/openapi.yaml` é a fonte; Go gerado (`std-http`, modo strict) e TypeScript gerado (tipos, SDK, hooks TanStack Query, Valibot) numa única pasta; CI regenera e falha com diff.

## Consequências
OpenAPI 3.0 (limite do gerador Go); `hey-api` ainda 0.x: versão exata e spike no Sprint 0.

## Alternativas descartadas
Huma (código-primeiro, OpenAPI 3.1); ogen; clientes escritos à mão.

## Fonte
`docs/fase-3/03-arquitetura-features.md; fase-3/openapi-sketch.yaml`
