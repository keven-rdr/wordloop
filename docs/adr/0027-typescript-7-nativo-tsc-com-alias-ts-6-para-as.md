# ADR 0027 — TypeScript 7 nativo (`tsc`) com alias TS 6 para as ferramentas

**Status:** aceita (2026-10-06)

## Contexto
O Sprint 0a mostrou que `typescript-eslint` e `hey-api` ainda exigem a API JS do TypeScript 6; o TS 7 não expõe API programática estável (prevista para o 7.1).

## Decisão
`@typescript/native` (`npm:typescript@7.0.2`) fornece o `tsc` 7 para tipagem e build; o pacote `typescript` é um alias de `@typescript/typescript6` para ferramentas que usam a API. Vite, Vitest e Biome não dependem do pacote. `skipLibCheck: true`.

## Consequências
Duas versões do compilador em `devDependencies` (o `tsc6` fica disponível); reavaliar a configuração quando o TS 7.1 trouxer a nova API; ADR 0013 mantém o StyleX 0.x com versões exatas.

## Alternativas descartadas
TypeScript 6 puro (perde o `tsc` nativo); TypeScript 7 puro (quebra lint e geração do cliente); trocar de ferramenta de lint.

## Fonte
`docs/sprint-0a/spike-report.md`
