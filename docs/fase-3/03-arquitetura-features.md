# Fase 3 — Funcionalidades, arquitetura, identidade e push

> **Decisões do usuário após a Fase 3 (prevalecem sobre o texto abaixo):** (1) **um único Keycloak** na VM, com **dois realms** (`wordloop-tst` e `wordloop-prd`), em vez de um por ambiente (a recomendação era um por ambiente); (2) BFF com cookie: sim; (3) pausa automática dos lembretes após 3 sem abertura: sim; (4) `page` começa em 1: ok.
> **Consequências do item 1:** economiza ~1,25–1,5 GB de RAM (total estimado ~3,5–4,5 GB sem observabilidade, o que volta a caber em 8 GB, embora você tenha escolhido 12 GB); **perde-se o ensaio isolado de atualização do IdP**: o tst e o prd passam a depender do mesmo processo. Mitigação (Fase 4): versão fixa por *digest*, exportar os dois realms antes de atualizar, ensaiar a atualização no Compose de **dev** com o export do realm, janela de manutenção avisada e Keycloak fora do fluxo de deploy dos apps.

> Data: 2026-10-06. Produto: **wordloop**. Rótulos: `[PREMISSA]` assumi; `[VERIFICAR]` não confirmei em fonte primária; `[DECISÃO SUA]` só você decide.
> **Arquivos desta fase:** [estrutura de pastas](03b-estrutura-pastas.md) · [diagramas C4 e sequências](03c-diagramas.md) · [design system StyleX](03d-design-system-stylex.md) · [DDL do estado do usuário](schema-learning.sql) · [esboço OpenAPI](openapi-sketch.yaml) (YAML validado por parser; **SQL e Mermaid não foram executados/renderizados**).
> Versões conferidas hoje no npm e no proxy do Go: Vite 8.3.3, Vitest 5.0.3, TypeScript 7.0.2, React 19.3.0, TanStack Router 1.170, React Router 8.4, i18next 26.4 / react-i18next 17.0, hey-api 0.99, vite-plugin-pwa 2.0.0, Go: pgx 5.11, sqlc 1.31, goose 3.28, oapi-codegen 2.8, huma 2.39, golangci-lint 2.14, go-fsrs v4.0.0, webpush-go 1.4.0.

## 1. Tabela de decisões

| Decisão | Recomendação | Alternativas descartadas | Por quê |
|---|---|---|---|
| Identidade | **Keycloak** (um por ambiente), a **API como BFF**: faz o OIDC *code + PKCE* e entrega só cookie `httpOnly` | auth própria em Go; Authentik/Zitadel; SPA com token no JS | Vínculo por e-mail, verificação, recuperação e MFA prontos e auditados; token nunca chega ao JavaScript |
| Contrato | **OpenAPI 3.0 primeiro** → `oapi-codegen` (servidor `std-http`, modo *strict*) e `hey-api` (cliente + hooks) | Huma (código-primeiro, OpenAPI 3.1); ogen | Você pediu contrato único; 3.0 porque o gerador Go é 3.0 `[VERIFICAR]` |
| HTTP | `net/http` (`ServeMux` com método+caminho) | chi, gin, echo | Zero dependência; `chi` v5.3 fica como plano B |
| Banco | **pgx + sqlc + goose** | GORM, ent | SQL visível e tipado; sem reflexão; migrações simples |
| Fila | **Outbox no Postgres** (`FOR UPDATE SKIP LOCKED`) | RabbitMQ; River (MPL-2.0, v0.49) | Atende ao volume; River é a evolução natural se o worker crescer |
| Eventos | Barramento **em processo**; `ReviewCompleted` roda **na mesma transação** | broker externo | Estatística consistente sem infraestrutura |
| Push | **Web Push nativo** (VAPID, `webpush-go`) | OneSignal; ntfy | Gratuito, sem terceiros nos dados, funciona no iOS instalado |
| Módulos | **5 + 1 biblioteca**: `identity`, `catalog`, `learning`, `progress`, `reminders` + `srs` pura | 6 módulos (com `scheduling`) | `scheduling` não tem I/O próprio: é biblioteca consumida por `learning` |
| Roteamento web | **TanStack Router** por arquivo | React Router 8 (rotas centrais); manifesto próprio | Rota = arquivo novo; árvore gerada, ninguém edita |
| i18n | `react-i18next` + `i18next-resources-to-backend`, tipos **gerados** por script | lista central em `i18n.ts` | Namespace novo = JSON novo |
| Paginação | **`page` começa em 1**, `size` (padrão 20, máx. 100), `sort=campo,asc` | 0-indexada (Contratos) | Escrito uma vez; casa com a UI |
| Erros | Problem Details (RFC 9457) com `code` + `params`; **o servidor devolve códigos, o front traduz** | texto localizado no servidor | Um catálogo de UI só; servidor só traduz o que ele mesmo envia (push) |
| E-mail | Brevo via SMTP do Keycloak (dev: Mailpit) | Resend, Mailjet, SES | Maior cota gratuita (300/dia) `[VERIFICAR]` |

## 2. Funcionalidades → implementação

| Funcionalidade | Como funciona |
|---|---|
| **Estudo** | "Quanto tempo hoje?" (5/15/30/60) → `POST /study/plan` devolve **pacote** (cartões + conteúdo + mídia); revisões vão em lote idempotente (UUID do cliente). Dificuldade **sempre** perguntada, com sugestão pré-selecionada por tempo (Fase 1) |
| **Exercícios** | `ExerciseGenerator` (Strategy, `supports`/`handle` como no Contratos) no servidor monta o enunciado e os distratores (mesma faixa de frequência, **sem irmãos semânticos**); no front um **registro de renderizadores** por tipo. A correção é refeita no servidor com a mesma função de normalização/distância de edição |
| **Metas** | `goal_cycle` semanal. Início e fim de ciclo perguntam de novo, **cumprida ou não**. 3 sugestões: leve = 0,8× sua mediana real dos últimos 14 dias, constante = mediana, intensa = percentil 75, limitadas a 5–120 min; sem histórico: 15/30/60. Meta semanal = dias de estudo (3/5/7); desafios de sequência 7/15/30 à parte |
| **Foco na melhora** | Retrospectiva do ciclo mostra **ganhos**: Δ palavras conhecidas, Δ acerto (7 d × 7 d anteriores), Δ tempo de resposta, itens que viraram "dominado". Dia estudado = **≥ 1 revisão** (minutos não contam para a sequência) |
| **Painel** | `study_day` pré-agregado (heatmap, acertos, sequência: no máximo 365 linhas lidas); "mais erradas" = `card.lapses` (índice parcial); previsão = contagem de `due_at` por dia lógico (índice); cobertura = Σ `coverage_share` dos lemas em `user_lexeme_known` |
| **Lembretes** | Seção 4 |
| **Configurações** | Idioma, fuso, tema, metas, lembretes, dispositivos, estratégia do agendador e retenção (avançado) |
| **Versão visível** | `GET /api/v1/version` + constantes injetadas no build do front → "v1.4.0 · API v1.7.2 · tst · a1b2c3d" |
| **LGPD** | `GET /me/export` (JSON) e `DELETE /me` (apaga dados do app e o usuário no Keycloak pela Admin API com conta de serviço de escopo mínimo). O gatilho de imutabilidade do `review_log` tem rotina própria de exclusão, registrada |

## 3. Identidade

| | A) Keycloak + BFF (recomendado) | B) Autenticação própria em Go |
|---|---|---|
| Vincular contas | fluxo *first broker login*: e-mail já existe → confirma por **e-mail** ou **reautenticação** ([resultado de busca](https://skycloak.io/blog/keycloak-account-already-exists-error/)) | escrever e testar à mão |
| Segurança | código de terceiros muito usado; **exige atualização constante** | você escreve argon2id, recuperação de senha, limites, *pre-hijacking* |
| RAM | **~1,25–1,5 GB por instância** (guias de dimensionamento; Authentik ≈ 0,6 GB ocioso, citado em blog `[VERIFICAR]`) | ~0 |
| Valor de aprendizado | OIDC, BFF, operação de IdP | alto, mas em área onde erro custa caro |
| Prazo | dias | semanas |

**Recomendação:** A. Com 12 GB sobra memória; o Contratos já usa Keycloak. **Um Keycloak por ambiente** (tst e prd), para testar atualização do IdP antes da produção; custo ≈ 2,5 GB.
**Por que BFF:** o navegador só guarda cookie `wl_session` (`httpOnly`, `Secure`, `SameSite=Lax`); o refresh token fica cifrado em `auth_session`. Mesma origem (`/api` atrás do Caddy) dispensa CORS. Custo: ~300 linhas de cliente OIDC (`go-oidc` v3.21 + `oauth2`) e proteção CSRF (cabeçalho custom + `SameSite`).**Segurança do vínculo** (ataque de *pre-hijacking*: [Sudhodanan & Paverd, USENIX Security 2022](https://www.usenix.org/conference/usenixsecurity22/presentation/sudhodanan)):
1. Só vincular com **prova de posse** (e-mail de confirmação ou senha da conta existente).
2. Cadastro por senha só fica utilizável **depois de verificar o e-mail**; ao vincular, descartar credenciais de conta nunca verificada `[VERIFICAR no teste de configuração]`.
3. Google: confiar em `email_verified`; GitHub: e-mail primário verificado (`user:email`); `Trust Email` ligado só no Google `[PREMISSA]`.
4. Senhas e MFA ficam no Keycloak (hash gerenciado por ele, não precisa de argon2id no nosso código).
**Perfil do app:** `app_user.id = sub`, criado no primeiro login (*JIT*); e-mail e senha não são duplicados no nosso banco.
**Tema do login:** começar com o tema padrão e cores/logo próprios; [Keycloakify](https://github.com/keycloakify/keycloakify) (tema em React) só se valer o esforço depois.

## 4. Lembretes por Web Push

| | Web Push nativo (recomendado) | OneSignal | ntfy |
|---|---|---|---|
| Custo | zero | web push gratuito, **10 mil destinatários por envio** na camada Free ([fonte secundária](https://onesignal.com/blog/understanding-onesignals-pricing)) | zero (auto-hospedado) |
| Privacidade | dados só no seu servidor | terceiro vê usuários e mensagens | no seu servidor |
| iOS | PWA **instalado** + permissão por gesto (≥ 16.4) ([resumo](https://pushpad.xyz/blog/ios-special-requirements-for-web-push-notifications)) | idem (usa o mesmo Web Push) | exige o PWA/app do ntfy: **não** é o seu app |
| Esforço | assinatura, VAPID, outbox, SW | SDK e painel | integração fraca com o app |

Base: [RFC 8030](https://www.rfc-editor.org/rfc/rfc8030.html) (protocolo), [RFC 8291](https://www.rfc-editor.org/rfc/rfc8291.html) (cifra do payload), [RFC 8292](https://www.rfc-editor.org/rfc/rfc8292.html) (VAPID). Biblioteca: [`webpush-go`](https://github.com/SherClockHolmes/webpush-go) v1.4.0 (MIT; último commit abr/2026). **Web Push × webhook:** Web Push é o servidor falando com o navegador do usuário por um serviço de push; webhook é um servidor avisando outro (ex.: GitHub avisando a VM de um deploy).

**Desenho** (sequência em [03c](03c-diagramas.md)):
- **Pedido de permissão** só depois de **1ª sessão concluída** e **por toque do usuário**; no iPhone, só depois de instalar o PWA (tela de instruções).
- **Assinatura por dispositivo** em `push_subscription`. 404/410 → `expired`.
- **Agendador** (worker): por usuário e dia, sorteia até `max_per_day` `reminder_slot` dentro da janela, fora do horário silencioso e nos dias escolhidos, respeitando o fuso; sorteio com semente `hash(user, data)` (testável).
- **Conteúdo:** a **pergunta curta** do cartão vencido/aprendendo de menor R; **nunca a resposta**. `hide_content` troca por "Você tem um desafio". `tag = notification.id` evita empilhar. Limite de ~4 KB do payload.
- **Deep link** `/study/challenge/{id}`; a resposta entra como `source = notification`.
- **Idempotência:** `UNIQUE (topic, event_key)` na outbox + `processed_event`; falha temporária → `attempts+1` e *backoff*; `notification_delivery` guarda o log de falhas.
- **Anti-fadiga:** máximo por dia, silêncio, **pausa automática** após 3 notificações seguidas sem abertura `[PREMISSA]`.

## 5. Arquitetura da API

**Módulos** (crítica à sua sugestão): `scheduling` deixa de ser módulo e vira a biblioteca pura `internal/srs`, porque não tem tabela nem caso de uso próprio (o estado do cartão pertence a `learning`). Forma interna de cada módulo: `domain/` · `app/` (um caso de uso por arquivo) · `adapters/`. Estrutura e regras impostas pelo CI: [03b](03b-estrutura-pastas.md).

| Padrão | Onde aplicar | Onde **não** aplicar |
|---|---|---|
| Strategy | `srs.Scheduler`; `ExerciseGenerator` e `Grader` por tipo; envio de notificação (`Sender`: justificado pelo teste com *fake*; 2º canal só quando existir) | provedores de login (o Keycloak já abstrai); provedores de imagem (só existem no pipeline do seed, não na API) |
| Repository | só como **funções do sqlc** por módulo | interface por tabela; `BaseRepository` |
| Unit of Work | `platform/db`: transação passada por `context`; `submit_reviews` grava log + cartão + `study_day` juntos | em leituras |
| Eventos de domínio | `ReviewCompleted` (síncrono, mesma transação) → `progress` | para falar com `identity` ou `catalog` (chamada direta por interface) |
| Registry/Factory | `srs` por nome (`simple_v1`, `fsrs6`); geradores de exercício por tipo | injeção por reflexão: usar construtores explícitos em `cmd/api` |

**Outros pontos:** um binário, dois papéis (`ROLES=http,worker`), dois contêineres da mesma imagem; API *stateless* (sessão no Postgres); `slog` em JSON, `/health/live`, `/health/ready`, `/metrics`; limite de taxa por usuário/IP em memória (`x/time/rate`, uma instância basta).
**i18n do servidor:** catálogos YAML **só para o que o servidor envia** (texto de push, mensagens de erro de depuração) com chaves tipadas geradas (`go generate`), lint contra string crua e teste que exige toda chave em `pt-BR` e `en-US`.
**Clientes offline:** o cliente corrige para dar feedback, mas o servidor **refaz a correção** e recalcula a nota; divergências ficam registradas. Duas revisões do mesmo cartão por dispositivos diferentes são aplicadas por `answered_at`, e ambas ficam no log.

## 6. Arquitetura do front

- **Rotas:** `src/routes/**` por arquivo (TanStack Router, divisão automática de código). Cada arquivo tem poucas linhas e importa a feature. O `routeTree.gen.ts` é gerado e marcado como gerado. React Router 8 foi descartado porque o modo framework centraliza em `routes.ts` e o modo de dados não gera a árvore.
- **Menu:** cada feature exporta `feature.ts`; o *shell* os descobre com `import.meta.glob`. Criar feature não edita arquivo central.
- **i18n:** `locales/{pt-BR,en-US}/<ns>.json`; `i18next-resources-to-backend` com `import(`../locales/${lng}/${ns}.json`)` (o Vite resolve o padrão); `useTranslation("study")` carrega o namespace sob demanda; `const { t: tCommon } = useTranslation("common")`. Tipos de chave: `scripts/gen-i18n-types` gera `i18next.d.ts` a partir dos JSON de `pt-BR`; o CI roda `check-i18n` (chaves iguais nos dois idiomas, sem string de UI fora de `t()`). Plural pelo CLDR do próprio i18next e `Intl` para datas/números; ICU só se faltar `[VERIFICAR]`.
- **API:** `src/api/generated/` (hey-api: tipos, SDK, hooks TanStack Query, schemas Valibot) é a **única** pasta gerada; o CI regenera e falha se houver diff. `hey-api` ainda é 0.x (0.99): versão exata e `[VERIFICAR]` no spike que os plugins de TanStack Query e Valibot cobrem o necessário. Erros: `useProblemHandler` mapeia `code` → `t("errors:<code>")` e `errors[].field` → `setError` do React Hook Form (modelo do `useErrorHandler` do Contratos).
- **Offline/PWA:** `vite-plugin-pwa` 2.0 em `injectManifest`; `src/sw.ts` (push, `notificationclick`, precache, atualização por aviso "nova versão"); fila de revisões e pacote de estudo em **IndexedDB**; envio ao reabrir e no evento `online` (Background Sync não existe no iOS).
- **Config em runtime:** `environment.json` validado por Valibot antes de montar; versão e commit via `define`.
- **Visual:** [03d](03d-design-system-stylex.md). Telas de exercício com **registro de renderizadores** (um módulo por tipo, descobertos por `import.meta.glob`). Gráficos: SVG próprio.

## 7. Dados

DDL completo do estado do usuário em [schema-learning.sql](schema-learning.sql) e do conteúdo em [content-schema.sql](../fase-2/content-schema.sql). Relações principais: `app_user` 1—N `card` N—1 `item`; `card` 1—N `review_log`; `review_log` → `study_day` (agregação na mesma transação); `notification` → `card`/`review_log`.
**Agregação:** `study_day` (uma linha por usuário/dia) alimenta heatmap, acertos e sequência; `known_lexemes` e `mastered_items` são gravados como **foto no fim do dia** (primeira escrita do dia seguinte fecha o dia anterior) `[PREMISSA]`, o que permite medir "melhora". Índices parciais cobrem fila de vencidos, mais erradas e previsão. **Retenção do log:** sem expurgo no MVP; particionar `review_log` por mês só se passar de dezenas de milhões de linhas.

## 8. Riscos, spikes do Sprint 0 e decisões

| Risco | Mitigação |
|---|---|
| ~~TypeScript 7.0.2 e ferramentas que usam a API JS~~ **Resolvido no Sprint 0a:** `typescript-eslint` e `hey-api` exigem a API do TS 6; usar `"typescript": "npm:@typescript/typescript6@^6.0.2"` + `"@typescript/native": "npm:typescript@7.0.2"` (`tsc` 7). Ver `docs/sprint-0a/spike-report.md` | convivência TS 6 (API) + TS 7 (`tsc`); reavaliar no TS 7.1 |
| StyleX + Vite 8 + Vitest 5 + React Compiler | spike (ver 03d) |
| `hey-api` 0.x muda de API | versão exata; geração isolada em uma pasta |
| `oapi-codegen` só em OpenAPI 3.0 | escrever 3.0.x; Huma é a saída se 3.1 virar necessidade |
| RAM: 2 Keycloaks + 2 ambientes | medir no spike com limite de heap da JVM `[VERIFICAR]`; 12 GB confortável |
| Vínculo de contas mal configurado | teste automatizado do fluxo no E2E contra o Keycloak de tst |
| SQL/Mermaid não validados | aplicar migração num Postgres local no Sprint 0; renderizar diagramas |

**Spikes (1–2 dias cada):** (1) Vite 8 + StyleX + Vitest + Base UI; (2) `oapi-codegen` std-http + `hey-api`; (3) Keycloak + BFF + vínculo Google/senha; (4) Web Push no Android, iPhone instalado e desktop; (5) migrações goose + sqlc.
**`[DECISÃO SUA]`:** (1) Keycloak por ambiente (≈ 2,5 GB) ou um só com dois *realms*? Recomendo um por ambiente. (2) BFF com cookie, mesmo sendo mais código que SPA com token? Recomendo BFF. (3) Pausa automática dos lembretes após 3 sem abertura: ok? (4) `page` começando em 1: ok?

## 9. Resumo de continuidade (≤ 300 palavras)

Fase 3 concluída (falta "ok" e 4 decisões da seção 8). Arquitetura: monólito modular Go com 5 módulos (`identity`, `catalog`, `learning`, `progress`, `reminders`) + biblioteca pura `srs`; forma `domain/app/adapters` por módulo; limites por CI (arquivo ≤ 300 linhas, função ≤ 40, ciclomática ≤ 10, ≤ 12 arquivos/pacote, `depguard` + `go-arch-lint`). Pilha: `net/http`, OpenAPI 3.0 → `oapi-codegen` (std-http strict), pgx + sqlc + goose, outbox Postgres (`SKIP LOCKED`), eventos em processo (`ReviewCompleted` na mesma transação). Identidade: Keycloak por ambiente + API como BFF (cookie `httpOnly`, refresh cifrado em `auth_session`), vínculo por e-mail ou reautenticação, e-mail verificado antes do uso, Brevo SMTP. Push: Web Push nativo (VAPID, `webpush-go`), slots sorteados por dia na janela do usuário, pergunta sem resposta, deep link, 404/410 expira, pausa após 3 sem abertura; iPhone só com PWA instalado. Front: React 19 + TS + Vite + StyleX; TanStack Router por arquivo; `feature.ts` por `import.meta.glob`; i18n por namespace sob demanda com tipos gerados; hey-api em `src/api/generated`; PWA `injectManifest`, IndexedDB offline com lote idempotente; Base UI + primitivas próprias; Astryx só como referência. Metas semanais dinâmicas com 3 sugestões (0,8× mediana / mediana / p75), retrospectiva de ganhos, dia estudado = ≥ 1 revisão. Dados: `schema-learning.sql` (card, review_log append-only, study_day, goal_cycle, push, outbox). Servidor refaz a correção e a nota. Riscos: TypeScript 7, StyleX 0.19.1 pré-1.0, hey-api 0.99, RAM. Próxima: Fase 4 (engenharia: repositório, pipelines, versão, ambientes, deploy, qualidade, kit para IA).
