# ADR 0016 — PWA offline com fila idempotente e pacote de estudo local

**Status:** aceita (2026-10-06)

## Contexto
Estudar sem rede; iOS não tem Background Sync.

## Decisão
`vite-plugin-pwa` (`injectManifest`); pacote de cartões e conteúdo em IndexedDB; revisões com UUID enviadas em lote ao reabrir ou no evento `online`; o servidor refaz correção e nota.

## Consequências
Conflitos entre dispositivos resolvidos por `answered_at`; mais testes.

## Alternativas descartadas
App só online; Background Sync (indisponível no iOS).

## Fonte
`docs/fase-3/03-arquitetura-features.md`
