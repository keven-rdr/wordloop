# ADR 0010 — Um Keycloak com dois realms (tst e prd)

**Status:** aceita (2026-10-06)

## Contexto
Identidade pronta (vínculo por e-mail, verificação, MFA); a VM tem 12 GB.

## Decisão
Uma instância do Keycloak (versão fixa por digest) com realms `wordloop-tst` e `wordloop-prd`; PostgreSQL próprio (`keycloak`).

## Consequências
Ponto único de falha para tst e prd e sem ensaio isolado de atualização: exportar realms antes, ensaiar em dev, janela de manutenção. Economiza ~1,25–1,5 GB.

## Alternativas descartadas
Um Keycloak por ambiente (recomendado, recusado pelo dono); autenticação própria em Go; Authentik/Zitadel.

## Fonte
`docs/fase-3/03-arquitetura-features.md (nota de decisões)`
