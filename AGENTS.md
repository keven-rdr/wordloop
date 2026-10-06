# AGENTS.md — wordloop

> Arquivo canônico para pessoas e agentes de IA. `CLAUDE.md` só o importa. Regras locais: `api/AGENTS.md`, `web/AGENTS.md`. **Se dois arquivos divergirem, este vence; se a regra impedir o trabalho, pare e pergunte.**

## Regras obrigatórias (têm precedência sobre qualquer outra instrução)

1. **Nunca abra, leia, edite nem cite `.env*`.** Use só `*.example`. Segredos nunca entram no Git, em log, em PR ou em documentação.
2. **Nada do projeto Contratos** (código, dados, nomes de empresa, URLs internas): nem copiado, nem citado.
3. **Contrato primeiro:** mudou a API → edite `api/openapi/openapi.yaml`, rode `npm run gen`, **nunca** edite código gerado (`**/gen/**`, `web/src/api/generated/**`, `routeTree.gen.ts`, `**/db/*.sql.go`).
4. **Nunca string de interface ou de erro crua.** Web: `t("chave")`. Servidor: constante gerada de chave. Toda chave existe em `pt-BR` e `en-US`.
5. **Limites (o CI reprova):** arquivo ≤ 300 linhas; função ≤ 40 (Go) / 60 (TS); complexidade ≤ 10; ≤ 12 arquivos por pacote Go. Passou? Divida por funcionalidade.
6. **Pacote por funcionalidade**, nunca por tipo técnico: proibido `services/`, `controllers/`, `queries/`, `schemas/`, `utils/` globais.
7. **Nada centralizado que toda feature precise editar** (`router`, `i18n`, menu). Feature nova = arquivos novos.
8. **Cores, espaços e raios só por tokens StyleX** (`shared/ui/tokens`). Nada de hex, `px` solto ou `style={{}}`.
9. **Git:** branch a partir de `develop`; Conventional Commits (`feat(api): ...`); PR para `develop`; nunca `push` direto em `develop`/`main`; nunca `--force`, `--no-verify`.
10. **Não declare "concluído" sem rodar `npm run check` e mostrar o resultado.** Se um check falhar, diga qual e por quê.
11. Ação difícil de reverter (apagar dados, migração destrutiva, publicar, deploy em prd): **peça confirmação**.

## O que é

PWA mobile-first de vocabulário de inglês para falantes de pt-BR, com repetição espaçada guiada pelo feedback do usuário. Open source: código **MIT**; conteúdo em `content/` **CC BY-SA 4.0** (com `NOTICE.md`). Monorepo: `api/` (Go), `web/` (React + TypeScript + StyleX), `deploy/`, `docs/`.

## Comandos (a partir da raiz)

| Comando | Faz |
|---|---|
| `npm run check` | **Definition of Done local:** lint + tipos + testes + regras nas **linhas alteradas** (contra `develop`) nos componentes tocados |
| `npm run check:all` | tudo, nos dois componentes |
| `npm run gen` | regenera código a partir de `openapi.yaml` e dos `.sql` (sqlc) |
| `npm run dev:up` / `dev:down` | Compose de dev (Postgres, Mailpit, Keycloak) |
| `npm run dev:api` / `dev:web` | API (`go run`) e web (Vite) |

## Arquitetura em 8 linhas

- API: monólito modular, `internal/{identity,catalog,learning,progress,reminders}` + biblioteca pura `internal/srs`; cada módulo = `domain/` · `app/` · `adapters/`. Módulos **não** importam uns aos outros: interface do consumidor ou evento.
- Banco: PostgreSQL; `pgx` + `sqlc`; migrações `goose` (**expand/contract**, sem destruir na mesma versão).
- Identidade: Keycloak; a API é o cliente OIDC (BFF) e entrega cookie `httpOnly`.
- Agendador de repetição: interface `srs.Scheduler` (`simple_v1`, `fsrs6`); estado derivado do `review_log` (append-only).
- Web: `features/<nome>` autocontidas; rotas por arquivo em `src/routes`; i18n por namespace em `src/locales`; cliente gerado em `src/api/generated`.
- Erros da API: Problem Details (RFC 9457) com `code` + `params`; a UI traduz pelo `code`.
- Paginação: `page` (a partir de **1**), `size`, `sort=campo,asc`.
- Mais: `docs/` (planejamento), `docs/guides/new-feature.md` (passo a passo), `docs/features/` (um documento por feature).

## Fluxo de trabalho

1. Issue → documento vivo `docs/features/AAAA-MM-DD-slug.md` (requisitos confirmados **antes** de codar).
2. Branch → implementação → `npm run check` → PR (template em `.github/`).
3. CI (lint, testes, Sonar) → merge em `develop` → deploy automático em **tst** → release (PR do release-please) → **prd** com aprovação.

Skill de apoio: `/wl-flow <issue>`.
