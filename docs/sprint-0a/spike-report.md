# Sprint 0a — relatório dos spikes de código

> Data: 2026-10-06. Ambiente: Windows 11, Node 24.14, npm 11.9, Docker 29.2.1, **Go 1.27.0** (instalado por `winget`; o `go.dev` já oferece 1.27.1). Os spikes de código estão **completos**; faltam só os de infraestrutura (0b). O disco `C:` está com ~9 GB livres.
> Código dos spikes: [web-stack/](web-stack/) · [contract/](contract/) · [go-stack/](go-stack/) · [postgres/](postgres/). Tudo foi executado de verdade; os números abaixo são saídas reais.

## 1. Resultado

| # | Spike | Parecer | Evidência |
|---|---|---|---|
| 1 | **PostgreSQL**: DDL de conteúdo + DDL de aprendizado + seed + conferência | **Aceito, em servidor real e em WASM** | **Servidor oficial `postgres:17-alpine` (17.8) em Docker:** os 2 DDLs, o seed (rodado 2×), a conferência (8 contagens `ok`) e [semantics.sql](postgres/semantics.sql) (idempotência, gatilho append-only, `UNIQUE`/`CHECK`, outbox com `SKIP LOCKED`, view de lemas conhecidos, índices parciais) passam com `ON_ERROR_STOP`. **Backup:** `pg_dump -Fc` (145 kB) restaurado com `pg_restore` em outro banco: mesmas contagens (51 itens, 194 textos, 1 revisão) e o gatilho continua bloqueando `UPDATE`. Também no motor PostgreSQL 18.3 (PGlite 0.5.8, WASM): Os 2 DDLs e o seed rodam; o seed é idempotente (51 itens, 194 textos, 34 lexemas, 10 mídias nas 2 execuções); 8 contagens `ok`, 9 conferências sem violação; gatilho do `review_log` bloqueia `UPDATE` e `DELETE`; reenvio do mesmo UUID não duplica; `UNIQUE`/`CHECK` de `card` e `outbox` funcionam; `FOR UPDATE SKIP LOCKED` roda |
| 2 | **Front**: Vite 8 + React 19 + StyleX + Vitest 5 + Base UI + TypeScript 7 | **Aceito, com ajustes** (seção 2) | `tsc` 7.0.2 sem erro; `vite build` em ~0,8 s (CSS 2,8 kB; JS 237 kB, 76 kB gzip); servidor de dev serve `/virtual:stylex.css` e transforma o código; 2 testes passam; Base UI (`Progress`, `Switch`) estilizado com StyleX |
| 3 | **React Compiler 1.0** com StyleX | **Aceito** (pode habilitar) | Build ok, CSS idêntico ao sem Compiler (mesmo hash), `useMemoCache` presente no bundle, 2 testes passam |
| 4 | **Lint**: ESLint + `@stylexjs/eslint-plugin` + Biome | **Aceito, com ajuste** | `valid-styles` e `no-unused` acusam erro em arquivo de teste negativo; `max-lines` e `complexity` ativos; `biome lint` sem erros |
| 5 | **Contrato**: OpenAPI → `hey-api` 0.99 (SDK, hooks TanStack Query, Valibot) | **Aceito, com ajuste** | 18 arquivos gerados em 0,4 s: 51 exports TanStack (`getVersionOptions`, `logoutMutation`…), schemas Valibot (`vProblem`, `vSettings`…); o código gerado passa no `tsc` 7 |
| 6 | **Validação do esboço OpenAPI** | **Corrigido** | `swagger-parser` achou erro real: `tags` globais precisam ser objetos. Corrigido em `docs/fase-3/openapi-sketch.yaml`; agora válido (3.0.3, 26 caminhos) |
| 7 | **Servidor Go a partir do contrato:** `oapi-codegen` v2.8.0 (`std-http`, *strict*) | **Aceito** | 2.761 linhas geradas; handler mínimo sobre `StrictServerInterface` + `net/http` (`ServeMux`): `GET /api/v1/version` → 200 com JSON; UUID inválido no caminho → 400; método errado → 405; `go vet` e 3 testes passam |
| 8 | **Migrações e consultas:** `goose` v3.28 + `sqlc` v1.31.1 + `pgx` v5.11 | **Aceito, com ajuste** | Os 2 DDLs viram migrações do goose (a função PL/pgSQL precisa de `-- +goose StatementBegin/End`); o `sqlc` lê as migrações e gera `ListDueCards`, `InsertReview`, `GetItemByNaturalKey`, `CountKnownLexemes` (inclui a *view*), com `uuid` mapeado para `google/uuid` |
| 9 | **Teste de integração:** Testcontainers (Go) v0.44 + `postgres:17-alpine` | **Aceito** | Em ~18–28 s: sobe o contêiner, aplica as 2 migrações, carrega o seed da amostra, roda as consultas do `sqlc`; 1º `InsertReview` insere, o 2º (mesmo UUID) afeta 0 linhas; `UPDATE` no log falha (append-only) |
| 10 | **Lint de arquitetura:** `golangci-lint` v2.14 (`depguard`, `funlen`, `cyclop`) e `go-arch-lint` v1.19 | **Aceito, com ajustes** | Limpo: 0 problemas / "No warnings". Negativos: `depguard` barra `pgx` na camada HTTP, `cyclop` acusa complexidade 12 > 10, `go-arch-lint` barra `api → learning` (aponta o arquivo e a linha) |

## 2. O que os spikes mudaram no plano

| Achado | Impacto | Ação |
|---|---|---|
| **`typescript-eslint` não suporta TS 7.0** (erro explícito); o `hey-api` também precisa do pacote `typescript` com API JS. A Microsoft documenta a convivência: `"typescript": "npm:@typescript/typescript6@^6.0.2"` (API 6) + `"@typescript/native": "npm:typescript@7.0.2"` (`tsc` 7) ([anúncio do TS 7](https://devblogs.microsoft.com/typescript/announcing-typescript-7-0/)) | O risco "TypeScript 7" da Fase 3 está **resolvido**, com esta configuração dupla | ADR 0027; `package.json` do spike como base do Sprint 1; reavaliar quando o TS 7.1 trouxer a API nova |
| **Media query no nível do objeto não existe no StyleX**: `[bp.reduceMotion]: { ... }` dentro de `stylex.create` dá "Invalid pseudo or at-rule". Condições vão **dentro do valor de cada propriedade**: `transitionDuration: { default: '120ms', [bp.reduceMotion]: '0s' }`, `boxShadow: { default: …, ':active': … }` | **Meu exemplo da Fase 3 estava errado** | Corrigido em `fase-3/03d-design-system-stylex.md`; regra em `web/AGENTS.md` |
| `@stylexjs/unplugin` declara `unplugin` como *peer* que **o npm não instala sozinho** | `vite build` e `vitest` falham ao carregar a config | `unplugin@2.3.11` exato no `package.json` |
| CSS do StyleX: em **produção** é anexado ao CSS que o Vite já emite (sem nenhum CSS, o `<link>` não aparece); em **dev** vem de `/virtual:stylex.css` + `virtual:stylex:runtime` | `import 'virtual:stylex.css'` quebra o build | `src/global.css` importado em `main.tsx`; em dev, `main.tsx` injeta o `<link>` e importa o *runtime* |
| `devMode: 'off'` **desliga a transformação** (erro "Unexpected stylex.defineVars call at runtime") | não serve para testes | usar `devMode: 'css-only'` no Vitest |
| **Vitest 5 + plugin do StyleX deixa 2 servidores Vite abertos** ("close timed out after 10000ms"): ~+10 s no fim da execução; **exit code 0**. Controle sem StyleX/jsdom não tem o problema; `server.watch: null` não resolve | CI mais lento em ~10 s e aviso no log | aceitar por ora; abrir *issue*/investigar no Sprint 1 |
| Tipos de terceiros ausentes (`@types/node`, `@types/babel__core`, `@types/babel__traverse`, `@types/aria-query`, `@testing-library/dom`) e conflito `Assertion` entre `@testing-library/jest-dom` e Vitest 5 | `tsc` com `skipLibCheck: false` falha só em `node_modules` | instalar os `@types`; `skipLibCheck: true`; importar `describe/it/expect` do `vitest` (sem `vitest/globals`) |
| `@vitejs/plugin-react` 6 + React Compiler exige `@rolldown/plugin-babel` **e** `@babel/core@7` explícitos | erro de módulo ausente | adicionados; `reactCompilerPreset()` testado |
| **`go-arch-lint`**: dependência vazia (`{}`) é rejeitada (exige `anyVendorDeps` ou `mayDependOn`); testes externos (`package x_test`) e código gerado geram falsos positivos; a **análise profunda** (`deepScan`) acusou `api → apigen` mesmo permitido | configuração do lint de arquitetura | `.go-arch-lint.yml` do spike: `excludeFiles` para `_test.go` e `gen/`, `deepScan: false` (só imports) |
| **`golangci-lint` v2**: formato novo (`version: "2"`, `linters.default: none`, `exclusions.rules`) | configuração da Fase 4 usa o formato antigo nos exemplos | usar o `.golangci.yml` do spike |
| **`sqlc` compila sem GCC** (usa WASM), mas o 1º `go run` leva ~5,7 min; `golangci-lint` via `go run` ~3 min | tempo de CI na 1ª execução | **instalar binários fixos** no CI (ou cache de módulos/build) em vez de `go run …@versão` |
| **Testcontainers**: precisa do daemon do Docker; no local usei `TESTCONTAINERS_RYUK_DISABLED=true` (evita baixar o contêiner "Ryuk") | CI do GitHub tem Docker; local depende de Docker Desktop ligado | documentar no `api/AGENTS.md`; manter Ryuk ligado no CI |
| `oapi-codegen` e `sqlc` via `//go:generate` com versão fixa; o código gerado **não** é versionado no spike | CI regenera e compara | decisão final (versionar ou não) no Sprint 1; hoje o plano diz "versionado e comparado" |
| `npm audit` (dev): 3 *high* em `braces` via `@stylexjs/eslint-plugin`; **0 em produção** | só ferramenta de lint | Renovate/`overrides` quando houver correção; não bloqueia |

## 3. O que não foi executado

| Spike | Motivo | O que preciso |
|---|---|---|
| Infraestrutura (0b): Keycloak + BFF, Web Push em Android/iPhone/desktop, release-please, deploy ponta a ponta, Sonar, `age` + R2 para o backup | dependem de VM, domínio e contas | compras e contas do Sprint 0b |
| `govulncheck` e `go-licenses` | não incluídos neste sprint | Sprint 1 (CI) |

**Versão do PostgreSQL:** validado em 17.8 (servidor real) e 18.3 (WASM); nenhum recurso exclusivo do 18 é usado. Para a VM, fixar **`postgres:17-alpine` por digest** `[PREMISSA]` (já validado) e reavaliar o 18 depois. **Limite:** o seed de amostra é pequeno; desempenho e volume só aparecem com o seed real.

## 4. Versões fixadas e validadas juntas (06/10/2026)

`react` 19.3.0 · `react-dom` 19.3.0 · `vite` 8.3.3 · `@vitejs/plugin-react` 6.1.2 · `@stylexjs/stylex`, `@stylexjs/unplugin`, `@stylexjs/eslint-plugin` 0.19.1 · `unplugin` 2.3.11 · `@base-ui/react` 1.8.0 · `vitest` 5.0.3 · `jsdom` 29.1.1 · `eslint` 10.12.0 · `@biomejs/biome` 2.5.15 · `@typescript/native` (`tsc` 7.0.2) · `typescript` (alias do 6.0.x) · `babel-plugin-react-compiler` 1.0.0 · `@babel/core` 7.29.7 · `@hey-api/openapi-ts` 0.99.0 · `valibot` 1.5.0 · `@tanstack/react-query` 5.104.1 · **Go** 1.27.0 · `oapi-codegen` 2.8.0 · `sqlc` 1.31.1 · `goose` 3.28.0 · `pgx` 5.11.0 · `testcontainers-go` 0.44.0 · `golangci-lint` 2.14.0 · `go-arch-lint` 1.19.0 · PostgreSQL 17.8 (Docker) e 18.3 (PGlite). O `package-lock.json` está em [web-stack/](web-stack/) e em [contract/](contract/).

## 5. Reprodução

```bash
cd docs/sprint-0a/web-stack && npm ci
npx tsc -b && npx vitest run && npx vite build && npx eslint src      # stack do front
npx vite build -c vite.config.rc.ts --outDir dist-rc                 # com React Compiler
cd ../contract && npm ci && npx openapi-ts                           # gera src/api/generated
cd ../go-stack && go generate ./... && go test ./... && go test -tags=integration ./internal/learning/...   # Go; integração exige Docker ligado
cd ../postgres && npm ci && node run.mjs                             # DDL + seed + semântica no PostgreSQL (PGlite)
# servidor real (Docker): ver os comandos `docker run --rm … postgres:17-alpine` e `psql -f` na seção 1; semantics.sql roda com psql
```
