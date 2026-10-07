# Criar uma feature (fatia vertical) — guia e feature de referência

Humanos e agentes seguem o **mesmo caminho**. A feature de referência é **Configurações do usuário** (`GET/PUT /me/settings`): é pequena e toca todas as camadas (contrato, banco, caso de uso, handler, hook gerado, formulário, i18n, rota, testes). Ao criar outra feature, copie a forma desta.

## Ordem

| # | Camada | O que fazer | Onde (exemplo da referência) |
|---|---|---|---|
| 1 | Documento | `docs/features/AAAA-MM-DD-slug.md` com definições confirmadas | `docs/features/…-user-settings.md` |
| 2 | Contrato | operação + schemas + erros em `api/openapi/openapi.yaml`; `npm run gen` | `getSettings`, `updateSettings`, schema `Settings` |
| 3 | Banco | migração **aditiva** em `api/migrations/`; consulta em `adapters/postgres/queries/*.sql`; `sqlc generate` | tabela `user_settings` |
| 4 | Domínio | tipo e validação em `internal/identity/domain/` (sem I/O) | `settings.go` |
| 5 | Caso de uso | um arquivo em `internal/identity/app/` | `update_settings.go` |
| 6 | Adaptador HTTP | handler da interface gerada em `adapters/http/`; erro = `problem.New(CodeXxx, params)` | `settings_handler.go` |
| 7 | Chaves de erro | catálogo do módulo → `go generate` (constantes tipadas) | `identity/messages/*.yaml` |
| 8 | Testes da API | unidade (caso de uso) + integração (handler + Postgres, `integration`) | `update_settings_test.go` |
| 9 | Web: rota | **arquivo novo** em `web/src/routes/` (poucas linhas) | `routes/settings.tsx` |
| 10 | Web: feature | `features/settings/{feature.ts,components,hooks}`; usa o **hook gerado** | `useGetSettings`, `useUpdateSettings` |
| 11 | Web: formulário | React Hook Form + Valibot; erro de API → `useProblemHandler` | `SettingsForm.tsx` |
| 12 | i18n | `locales/{pt-BR,en-US}/settings.json` novo; `npm run gen:i18n` | `useTranslation("settings")` |
| 13 | Estilo | `stylex.create` só com tokens; mobile-first; alvo ≥ 48 px | `SettingsForm.styles.ts` |
| 14 | Testes web | Vitest (componente) + Playwright (`e2e/settings/` com mocks de API) | `settings.spec.ts` |
| 15 | Fechar | `npm run check`; PR com o template | — |

## O que **não** se edita
Nenhum arquivo central: sem `router.tsx`, `i18n.ts` nem lista de menu. A rota é arquivo novo (a árvore é gerada), o menu vem do `feature.ts` (descoberto por `import.meta.glob`) e o namespace de i18n é carregado sob demanda. Se parecer que precisa editar um arquivo central, **pare**: a regra 7 do `AGENTS.md` foi violada.

## Armadilhas
- Corrigir resposta, nota e agendamento **só no servidor**; o cliente só dá feedback.
- Nunca `UPDATE`/`DELETE` em `review_log`.
- Migração que remove ou renomeia coluna só em versão posterior (expand/contract).
