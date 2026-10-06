# ADR 0003 — `srs` como biblioteca pura com Strategy (`simple_v1`, `fsrs6`)

**Status:** aceita (2026-10-06)

## Contexto
Quero entender e ajustar o algoritmo e poder compará-lo ao FSRS.

## Decisão
Interface `Scheduler{Next, Retrievability}` sem I/O; estratégias registradas por nome; escolha por configuração do usuário. `simple_v1` no MVP, `fsrs6` (go-fsrs v4, MIT) depois.

## Consequências
Testes de propriedade e simulação sem banco; trocar de estratégia = reprocessar o log (ADR 0005).

## Alternativas descartadas
Só FSRS (menos didático); só regra simples (pior retenção em dados reais, benchmark da comunidade).

## Fonte
`docs/fase-1/01-srs.md`
