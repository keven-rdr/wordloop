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
- SQL em `adapters/postgres/queries/*.sql` → `sqlc generate`. Migração nova = arquivo novo em `migrations/` (goose), **expand/contract**.
- Transação por Unit of Work (`platform/db`). `review_log` é **append-only**: nunca `UPDATE`/`DELETE`; correção é evento `review_void`.

## API
- Handlers implementam a interface *strict* gerada; **não edite `gen/`**. Mudou o contrato → `openapi.yaml` → `npm run gen`.
- Erro de API: `problem.New(CodeXxx, params)` com **constante gerada**; nunca string literal.
- Envio de revisão é **idempotente** pelo UUID do cliente; o servidor refaz a correção e a nota.

## Testes
- Todo caso de uso tem teste de unidade; handlers e repositórios têm teste de integração (Testcontainers, tag `integration`).
- `internal/srs`: golden vectors + testes de propriedade (`rapid`) + simulação (`cmd/srs-sim`).
- Comando: `cd api && go test ./...` (unidade) · `go test -tags=integration ./...`.

## Limites
Arquivo ≤ 300 linhas · função ≤ 40 · ciclomática ≤ 10 · ≤ 12 arquivos/pacote · módulo ≤ 60 arquivos (acima: dividir por funcionalidade).
