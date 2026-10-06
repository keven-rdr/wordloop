# ADR 0013 — StyleX com Base UI e tokens próprios

**Status:** aceita (2026-10-06)

## Contexto
Decisão de front já tomada: React + TypeScript + StyleX (0.x).

## Decisão
Tokens `defineVars`/`defineConsts`, temas `createTheme` (claro/escuro/sistema), primitivas Base UI sem estilo + componentes próprios; versões `@stylexjs/*` exatas e atualizadas juntas; ESLint só para StyleX e regras de tamanho.

## Consequências
Ecossistema pequeno; risco de quebra pré-1.0; spike com Vite 8/Vitest 5/React Compiler.

## Alternativas descartadas
Astryx (beta, só referência por ora); React Aria; componentes 100% próprios.

## Fonte
`docs/fase-3/03d-design-system-stylex.md`
