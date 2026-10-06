# ADR 0011 — API como BFF OIDC com cookie `httpOnly`

**Status:** aceita (2026-10-06)

## Contexto
Evitar tokens no JavaScript (XSS) e simplificar o PWA em mesma origem.

## Decisão
A API faz o fluxo code + PKCE; sessão no PostgreSQL (`auth_session`) com refresh token cifrado; cookie `httpOnly`, `Secure`, `SameSite=Lax`; proteção CSRF por cabeçalho.

## Consequências
~300 linhas de cliente OIDC e sessão por manter; a API deixa de ser só validadora de JWT.

## Alternativas descartadas
SPA com token em memória; JWT validado na API sem sessão.

## Fonte
`docs/fase-3/03-arquitetura-features.md; fase-3/03c-diagramas.md`
