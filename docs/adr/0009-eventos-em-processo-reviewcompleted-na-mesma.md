# ADR 0009 — Eventos em processo; `ReviewCompleted` na mesma transação

**Status:** aceita (2026-10-06)

## Contexto
Estatísticas (`study_day`) precisam ser consistentes com o log.

## Decisão
Barramento síncrono em memória; handler de `progress` roda antes do commit; integrações externas passam pela outbox.

## Consequências
Acoplamento temporal entre módulos dentro da transação; sem entrega a sistemas externos por este caminho.

## Alternativas descartadas
Broker externo; eventos assíncronos para tudo.

## Fonte
`docs/fase-3/03-arquitetura-features.md`
