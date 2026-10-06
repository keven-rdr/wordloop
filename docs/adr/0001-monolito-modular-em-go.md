# ADR 0001 — Monólito modular em Go

**Status:** aceita (2026-10-06)

## Contexto
Projeto pessoal, uma pessoa, sem orçamento; a API deve nascer genérica (outros assuntos no futuro).

## Decisão
Um binário Go com módulos `identity`, `catalog`, `learning`, `progress`, `reminders` e biblioteca pura `srs`; cada módulo com `domain/app/adapters`. Um binário, dois papéis (`http`, `worker`).

## Consequências
Deploy simples e depuração fácil; fronteiras precisam ser impostas por ferramenta (ADR 0002). Dividir um módulo quando passar de 60 arquivos.

## Alternativas descartadas
Microsserviços (custo operacional sem ganho); módulo único sem fronteiras (repetiria o `contracts` de 326 arquivos do Contratos).

## Fonte
`docs/fase-3/03-arquitetura-features.md`
