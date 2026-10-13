# Fundação do banco (Postgres, migrações e readiness)

> Arquivo: `docs/features/2026-10-13-fundacao-banco.md` · Issue: #__ · Branch: `feat/fundacao-banco` · Estado: BUILD

## Contexto
Sprint 1 (esqueleto ponta a ponta). A API ainda não fala com o banco. Esta fatia é a base do login (BFF) e de todos os módulos: conexão, migrações e prova de vida do banco. Sem tela nova.

## Definições confirmadas
- **Banco real em todos os ambientes, nada simulado:** dev (Compose), tst (Neon, sa-east-1) e prd. A API **exige** `DATABASE_URL` e falha ao iniciar sem ela.
- `pgx` v5 (pool), `goose` (migrações SQL embutidas no binário), sem ORM (ADR 0007).
- `GET /health/ready` passa a checar o banco: 200 se responde, **503** se não.
- Migrações rodam no deploy, **antes** de subir a versão nova: `api migrate` (expand/contract).
- Primeira migração: `app_user` (perfil mínimo; `id` = `sub` do Keycloak).
- Teste de integração com Postgres real (Testcontainers, tag `integration`), também no CI.

## Decisões
- `sqlc` entra junto com a primeira consulta (login/`upsert app_user`), não nesta fatia: sem consulta, não há o que gerar.
- Conexão com `MaxConns` baixo (5): o Neon free tem limite de conexões e a VM tem 1 GB.
- `Server` recebe uma interface `Pinger` declarada no pacote `api` (consumidor define a interface).

## Checklist de progresso
- [x] Contrato: `getReadiness` com 503 (`openapi.yaml` + `npm run gen`)
- [x] API: `platform/db` (pool + migrar), migração `app_user`, `api migrate`, readiness com banco, testes (unidade + integração)
- [x] Deploy: `deploy.sh` roda `migrate` antes do `up` (nunca no rollback); CI com job `test-integration`
- [ ] tst: Neon criado e `DATABASE_URL` em `secrets.env` na VM
- [ ] `npm run check` verde

## Verificação
Dev: `docker compose -f deploy/dev/compose.yaml up -d postgres`, `cd api && go run ./cmd/api migrate` (com `DATABASE_URL`), `curl localhost:8080/api/v1/health/ready` → 200; parar o Postgres → 503 (verificado). tst: `curl https://tst.wordloop.com.br/api/v1/health/ready` → 200 com o Neon.
