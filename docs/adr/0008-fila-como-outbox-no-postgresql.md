# ADR 0008 — Fila como outbox no PostgreSQL

**Status:** aceita (2026-10-06)

## Contexto
Envio de push e processamentos assíncronos com idempotência, sem nova infraestrutura.

## Decisão
Tabela `outbox` com `UNIQUE(topic,event_key)`, worker com `FOR UPDATE SKIP LOCKED`, *backoff*, `processed_event` por consumidor.

## Consequências
Throughput limitado ao do Postgres (suficiente); migrar para River (MPL-2.0) se o worker crescer.

## Alternativas descartadas
RabbitMQ (uma peça a mais para operar); River já agora.

## Fonte
`docs/fase-3/03-arquitetura-features.md`
