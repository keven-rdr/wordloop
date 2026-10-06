# ADR 0014 — TanStack Router por arquivo e `feature.ts` por `import.meta.glob`

**Status:** aceita (2026-10-06)

## Contexto
Criar feature no Contratos exige editar `router.tsx`, `i18n.ts` e `nav-items.ts`.

## Decisão
Rotas em `src/routes/` (árvore gerada); menu e namespaces declarados no `feature.ts` da feature; nada central editado à mão.

## Consequências
Arquivo `routeTree.gen.ts` gerado e versionado; rotas finas que importam a feature.

## Alternativas descartadas
React Router 8 (rotas centrais); manifesto próprio sem roteador de arquivos.

## Fonte
`docs/fase-3/03-arquitetura-features.md`
