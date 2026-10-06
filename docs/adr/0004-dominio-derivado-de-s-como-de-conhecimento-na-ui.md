# ADR 0004 — Domínio (derivado de S) como '% de conhecimento' na UI; R só para fila e notificação

**Status:** aceita (2026-10-06)

## Contexto
Mostrar R por item cairia sem o usuário agir e desmotivaria; mas R é a probabilidade com sentido estatístico.

## Decisão
Domínio = ln(1+S)/ln(181), alterado só ao responder; R = (1+F·t/S)^-0,1542 calculado na hora; uma curva compartilhada por todas as estratégias. 'Dominado' = S ≥ 90 d.

## Consequências
Dois números a explicar; calibrar S_max = 180 d com dados.

## Alternativas descartadas
Só R; escore próprio sem modelo de memória.

## Fonte
`docs/fase-1/01-srs.md`
