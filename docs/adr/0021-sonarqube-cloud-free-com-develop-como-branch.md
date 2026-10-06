# ADR 0021 — SonarQube Cloud Free com `develop` como branch principal

**Status:** aceita (2026-10-06)

## Contexto
O plano Free analisa PR e branch só da branch principal e não personaliza o quality gate.

## Decisão
Dois projetos (`api`, `web`) no Sonar; `develop` como principal; token só por secret; o que o Sonar não cobre é imposto pelo CI e pelo hook.

## Consequências
PRs de release (`develop → main`) não são analisadas.

## Alternativas descartadas
SonarQube self-hosted (sem análise de branch/PR e pesado); um projeto único.

## Fonte
`docs/fase-4/04-engenharia.md`
