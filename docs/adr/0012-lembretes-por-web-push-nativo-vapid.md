# ADR 0012 — Lembretes por Web Push nativo (VAPID)

**Status:** aceita (2026-10-06)

## Contexto
Notificação-desafio no celular e no PC, sem custo e sem terceiros nos dados.

## Decisão
`webpush-go` (RFC 8030/8291/8292); assinatura por dispositivo; slots sorteados por dia na janela do usuário; pergunta sem resposta; 404/410 expira; pausa após 3 sem abertura. iPhone só com PWA instalado.

## Consequências
Código próprio de assinatura/envio; limite de ~4 KB no payload; iOS exige instalação e gesto.

## Alternativas descartadas
OneSignal (terceiro vê usuários e mensagens); ntfy (não é o app do projeto).

## Fonte
`docs/fase-3/03-arquitetura-features.md`
