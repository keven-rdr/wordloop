# Fase 3 — Estrutura de pastas e regras de dependência

## `api/` (Go)

```
api/
├── cmd/
│   ├── api/main.go            # composition root: monta módulos e injeta dependências (ROLES=http,worker)
│   └── srs-sim/main.go        # simulador de alunos virtuais (Fase 1, seção 7)
├── openapi/openapi.yaml       # FONTE DA VERDADE do contrato
├── migrations/                # goose: 0001_init.sql ...
├── sqlc.yaml
└── internal/
    ├── platform/              # transversal, SEM regra de negócio
    │   ├── config/  db/ (pool, UoW)  httpx/ (middlewares)  problem/ (RFC 9457)
    │   ├── i18n/ (catálogos + chaves geradas)  version/  clock/  eventbus/  outbox/
    ├── srs/                   # BIBLIOTECA PURA (sem I/O): Scheduler, curva R(t,S), simple_v1, fsrs6
    │   ├── scheduler.go  curve.go  simple.go  fsrs.go  fuzz.go  testdata/ (golden vectors)
    ├── identity/
    │   ├── domain/            # User, Settings
    │   ├── app/               # login_callback.go, delete_account.go, export_data.go (1 caso de uso por arquivo)
    │   └── adapters/          # oidc/ (go-oidc), postgres/ (sqlc), http/ (handlers)
    ├── catalog/               # domain/ app/ adapters/ (mesma forma)
    ├── learning/              # plan_session.go, submit_reviews.go, get_challenge.go, grading/
    ├── progress/              # on_review_completed.go, goals.go, dashboard.go, coverage.go
    └── reminders/             # plan_slots.go, send_push.go, subscribe.go, prefs.go, webpush/ (adapter)
```
Código gerado fica em `internal/<modulo>/adapters/postgres/db/` (sqlc) e `internal/api/gen/` (oapi-codegen); nunca editado à mão, regenerado no CI e comparado (`git diff --exit-code`).

### Regras de dependência (impostas no CI)

| Regra | Ferramenta |
|---|---|
| `domain` não importa `app`, `adapters` nem `platform/db` | `depguard` (golangci-lint) |
| `app` importa só `domain`, `srs`, `platform` e **interfaces que ela mesma declara** | `depguard` |
| Um módulo **não importa o pacote de outro**; fala por interface do consumidor ou por evento | `go-arch-lint` (config por módulo) |
| `srs` não importa nada do projeto | `go-arch-lint` |
| Só `cmd/api` conhece implementações concretas (composition root) | `go-arch-lint` |

### Limites mensuráveis (o CI reprova)

| Métrica | Limite | Como |
|---|---|---|
| Linhas por arquivo (`.go`, sem gerados) | ≤ 300 | script `scripts/check-structure` |
| Linhas por função | ≤ 40 | `funlen` |
| Complexidade ciclomática | ≤ 10 | `cyclop` |
| Arquivos por pacote | ≤ 12 | `scripts/check-structure` |
| Arquivos por módulo | ≤ 60 → dividir | `scripts/check-structure` |

**Quando dividir um módulo:** passou de 60 arquivos, ou duas equipes/features mudam em ritmos diferentes, ou há duas linguagens de negócio (ex.: `learning` com sessões vs. *placement*). O sinal mais barato é um pacote com > 12 arquivos: vira subpacote por funcionalidade (`learning/grading`).
**Interfaces:** só onde há 2+ implementações (`srs.Scheduler`, canais de notificação, provedores de imagem) ou para isolar I/O em teste. Nada de `BaseRepository` nem par `IService`/`Service`.

## `web/` (React + TypeScript + StyleX)

```
web/
├── public/                     # manifest, ícones, environment.json (montado em runtime pelo Compose)
├── src/
│   ├── main.tsx
│   ├── core/                   # config (valida environment.json), providers, i18n, version, pwa (sw.ts), auth
│   ├── api/generated/          # ÚNICA pasta de código gerado (hey-api: tipos, SDK, hooks TanStack Query, schemas Valibot)
│   ├── shared/
│   │   ├── ui/                 # primitivas: tokens/*.stylex.ts, Button, Card, Progress, Dialog, Tabs, Toast, Input, Switch, Heatmap
│   │   ├── hooks/  lib/
│   ├── features/
│   │   ├── auth/  onboarding/  study/  dashboard/  goals/  settings/  library/  reading/
│   │   │   ├── feature.ts      # manifesto: { id, nav?, i18nNamespaces }  ← descoberto por import.meta.glob
│   │   │   ├── components/  hooks/  exercises/ (só study: um módulo por tipo de exercício)
│   ├── routes/                 # TanStack Router (por arquivo); cada rota tem ~5 linhas e importa a feature
│   ├── locales/{pt-BR,en-US}/<ns>.json    # common.json + um por feature; carregados sob demanda
│   └── sw.ts                   # service worker (push, notificationclick, precache, fila offline)
├── scripts/                    # gen-i18n-types, check-i18n, check-structure
└── e2e/                        # Playwright: specs por feature + mocks
```

### Regras do front

| Regra | Como se impõe |
|---|---|
| Feature não importa outra feature; só `shared/`, `core/` e `api/generated/` | ESLint `no-restricted-imports` (mesmo ESLint do StyleX) |
| Arquivo ≤ 300 linhas; função ≤ 60; complexidade ≤ 10 | ESLint `max-lines`, `max-lines-per-function`, `complexity` + Biome `noExcessiveCognitiveComplexity` |
| Sem `services/`, `queries/`, `schemas/` globais | revisão + `check-structure` (proíbe pastas com esses nomes na raiz de `src`) |
| Criar uma feature **não edita** arquivo central | rota = arquivo novo em `routes/` (o plugin regenera `routeTree.gen.ts`); menu = `feature.ts`; i18n = `locales/*/<ns>.json` novo; tipos de i18n = script gera |
| Nada de string de UI fora do `t()` | regra de lint + hook do Claude Code (Fase 4) |
