# ADR 0019 — Deploy por SSH com comando forçado e GitHub Environments

**Status:** aceita (2026-10-06)

## Contexto
VM própria; repositório público torna *runner* self-hosted arriscado.

## Decisão
Workflow reutilizável chama `deploy.sh` por chave com `command=`; Environment `prd` com revisor obrigatório e restrição a tags; rollback = redeploy da tag anterior; migrações expand/contract.

## Consequências
SSH exposto (só chave, `fail2ban`); estado das tags fica na VM e no histórico de deploys do GitHub.

## Alternativas descartadas
Runner self-hosted; commit de bot com tag no overlay; k3s + Argo CD.

## Fonte
`docs/fase-4/04-engenharia.md; fase-4/deploy/deploy-skeleton.md`
