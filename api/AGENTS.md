# api/ — regras locais (Go)

Vale junto com o `AGENTS.md` da raiz (que vence em conflito).

## Estrutura e dependências
- `internal/<modulo>/{domain,app,adapters}`; **um caso de uso por arquivo** em `app/` (`submit_reviews.go`).
- `domain` não importa `app`/`adapters`; `app` importa só `domain`, `srs`, `platform` e interfaces **que ela mesma declara**.
- Módulo não importa pacote de outro módulo. Fale por interface do consumidor ou por evento (`ReviewCompleted`).
- Só `cmd/api` conhece implementações concretas (composition root; injeção explícita, sem reflexão).
- `internal/srs` não faz I/O e não importa nada do projeto.

## Go idiomático
- Sem `BaseRepository`, `BaseEntity`, par `IService`/`Service`. Interface só com 2+ implementações ou para isolar I/O em teste.
- Erros: `fmt.Errorf("...: %w", err)`; erro de domínio é variável/tipo do pacote; HTTP converte em Problem Details com `code`.
- `context.Context` primeiro parâmetro; relógio injetado (`platform/clock`), nunca `time.Now()` em regra de negócio.
- Logs com `slog`, estruturados; nunca logar token, cookie, e-mail nem `answer_raw`.

## Banco
- **Obrigatorio em todos os ambientes:** a API exige `DATABASE_URL` e nao inicia sem ela; nada de banco simulado. Migracoes: `api migrate` (goose, embutido em `api/migrations/`); o deploy as roda antes de subir a versao nova.
- SQL em `adapters/postgres/queries/*.sql` → `sqlc generate`. Migração nova = arquivo novo em `migrations/` (goose), **expand/contract**.
- **Migração com muitas instruções (seed):** o `goose` manda uma instrução por vez e cada uma é uma ida e volta até o banco (~0,15 s até o Neon em tst). Um seed de ~850 instruções levou 2 min e estourou o limite do deploy. Envolva o bloco todo em `-- +goose StatementBegin` … `-- +goose StatementEnd`: vira **um** envio. Migração já aplicada não se edita; vale para a próxima (seed v1, ~500 palavras).
- Transação por Unit of Work (`platform/db`). `review_log` é **append-only**: nunca `UPDATE`/`DELETE`; correção é evento `review_void`.

## API
- Handlers implementam a interface *strict* gerada; **não edite `gen/`**. Mudou o contrato → `openapi.yaml` → `npm run gen`.
- Documentação: `GET /api/v1/docs/` (Swagger UI embutido, sem CDN) e `GET /api/v1/openapi.yaml` (o `openapi.yaml` embutido por `api/openapi/embed.go`). Ligada por padrão; `API_DOCS_ENABLED=false` desliga (404). Não é rota de negócio: não entra no contrato.
- **Quem faz o pedido vem só da sessão** (`identity/app.Authenticate`): nunca leia papel ou id de usuário do corpo, da query ou de cabeçalho. Papel = `app_user.role`, só muda por SQL do dono. Rota que lê/muda dado de usuário filtra pelo `Principal` (anti-IDOR: dado alheio responde 404). `DELETE`/`POST` que mudam estado exigem `X-Requested-With`. Admin só entra com 2 fatores (`amr`) e a sessão dele é curta (`domain.SessionPolicy`).
- Passkeys: a biblioteca WebAuthn fica só em `identity/adapters/webauthn`; casos de uso em `identity/passkeys` falam com a interface `Ceremony`. Desafio de **uso único e no servidor** (`auth_challenge`); convite consumido na **mesma transação** que grava a passkey; `userVerification: required` sempre. Toda forma de login abre sessão por `app.SessionOpener` (a regra "conta que ligou o 2FA só entra com 2 fatores comprovados" mora lá).
- Senhas: casos de uso em `identity/passwords`; o hash só passa por `platform/password` (Argon2id, no máximo 2 ao mesmo tempo). A senha nunca vai para log, URL, erro ou resposta. Toda recusa de login é o **mesmo 401** (hash isca quando a conta não existe) e 5 erros seguidos bloqueiam a conta por tempo crescente. Senha sozinha grava `MFA=false`. Ver ADR 0029.
- Erro de API: `problem.New(CodeXxx, params)` com **constante gerada**; nunca string literal.
- Envio de revisão é **idempotente** pelo UUID do cliente; o servidor refaz a correção e a nota.

## Testes
- Todo caso de uso tem teste de unidade; handlers e repositórios têm teste de integração (Testcontainers, tag `integration`).
- `internal/srs`: golden vectors + testes de propriedade (`rapid`) + simulação (`cmd/srs-sim`).
- Comando: `cd api && go test ./...` (unidade) · `go test -tags=integration ./...` (exige **Docker ligado**; localmente use `TESTCONTAINERS_RYUK_DISABLED=true`).
- Lint: `golangci-lint` v2 com `.golangci.yml` e `.go-arch-lint.yml` (`deepScan: false`, `_test.go` e `gen/` excluídos), ambos validados no Sprint 0a (`docs/sprint-0a/go-stack`).
- Geração: `go generate ./...` (`oapi-codegen` e `sqlc` com versão fixa). Migração com função PL/pgSQL: envolver em `-- +goose StatementBegin/End`.

## Limites
Arquivo ≤ 300 linhas · função ≤ 40 · ciclomática ≤ 10 · ≤ 12 arquivos/pacote · módulo ≤ 60 arquivos (acima: dividir por funcionalidade).
