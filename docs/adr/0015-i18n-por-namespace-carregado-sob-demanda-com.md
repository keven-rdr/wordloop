# ADR 0015 — i18n por namespace carregado sob demanda, com tipos gerados

**Status:** aceita (2026-10-06)

## Contexto
Textos de interface em pt-BR e en-US; novos idiomas sem mexer em código.

## Decisão
`react-i18next` + `i18next-resources-to-backend`; `locales/<lng>/<ns>.json`; tipos de chave gerados por script; `check:i18n` no CI; hook bloqueia texto literal. O servidor devolve códigos de erro; o front traduz.

## Consequências
Script de geração e checagem a manter.

## Alternativas descartadas
Lista central de namespaces; tradução de erros no servidor.

## Fonte
`docs/fase-3/03-arquitetura-features.md`
