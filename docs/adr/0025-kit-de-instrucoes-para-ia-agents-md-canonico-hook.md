# ADR 0025 — Kit de instruções para IA: `AGENTS.md` canônico, hook e skill

**Status:** aceita (2026-10-06)

## Contexto
No Contratos há três arquivos de instrução sobrepostos e já divergentes.

## Decisão
`AGENTS.md` (raiz, `api/`, `web/`) com regras obrigatórias e precedência; `CLAUDE.md` só importa; hook `PreToolUse` bloqueia `.env*`, código gerado, arquivos grandes, texto de UI e cor crua; skill `wl-flow` com gates; PR template e documento vivo.

## Consequências
Instrução é contexto, não garantia: o que não pode acontecer vai para hook e CI.

## Alternativas descartadas
Três arquivos paralelos; só revisão humana.

## Fonte
`docs/fase-4/04-engenharia.md; fase-4/kit-ia/`
