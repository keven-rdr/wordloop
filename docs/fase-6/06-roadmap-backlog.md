# Fase 6 — Roadmap, backlog do MVP, seed e fechamento

> Data: 2026-10-06. Produto: **wordloop**. Rótulos: `[PREMISSA]` assumi; `[VERIFICAR]` não confirmei; `[DECISÃO SUA]` só você decide.
> **Arquivos:** [seed: amostra + plano](seed/README.md) (gerador, SQL, conferência: **validados em SQLite, não em PostgreSQL**) · [ADRs](../adr/README.md) (26) · [referências auditadas](../referencias.md) · fases anteriores em `docs/fase-*`.
> **Estimativas de esforço são palpites meus**, sem histórico de entrega; trate como ordem de grandeza.

## 1. Resumo executivo

**wordloop** é um PWA mobile-first, open source, para falantes de pt-BR ampliarem o vocabulário de inglês com **repetição espaçada guiada pelo feedback** (Difícil/Médio/Fácil depois de cada resposta), **imagens como contexto depois da tentativa** e **notificações-desafio** que o usuário abre e responde. O progresso é medido por **palavras conhecidas e cobertura de textos**, e a meta é **dinâmica** (perguntada a cada ciclo). A trilha vai de palavras a frases, até **ler uma página** com ≥ 95% de cobertura. A API nasce **genérica** (`subject → course → unit → item`) para outros assuntos.
**Stack:** Go (monólito modular) + PostgreSQL + Keycloak · React 19 + TypeScript + Vite + StyleX · GitHub Actions, Sonar Cloud, Compose na VM Contabo (tst e prd).

| Decisão desta fase | Recomendação | Alternativa descartada | Por quê |
|---|---|---|---|
| Cadência | Sprints de **2 semanas**, ~16–20 h cada (8–10 h/semana `[PREMISSA]`) | Kanban sem sprints | Critério de aceite e marco visíveis para quem trabalha sozinho |
| Ordem | **Esqueleto ponta a ponta primeiro** (CI, deploy, login, versão), depois SRS, depois lembretes | Construir por camada | Pipeline e versão visíveis desde o Sprint 1; risco técnico cai cedo |
| MVP | = **Must + o que você pediu** (TTS, offline, nivelamento, lembretes), entregue em **incrementos** que já vão ao prd | Um único lançamento | Cada sprint termina em algo usável; o escopo é grande para uma pessoa |
| Conteúdo do MVP | ~**500 palavras** (T0 + parte de T1), ~100 frases, imagem nas concretas | 1.000 palavras | Revisão de tradução e curadoria de imagem são o gargalo real |
| Versão | 0.x até o MVP; primeiro release em prd = `1.0.0` por `release-as` | 1.0 só no fim sem prd antes | Permite *beta* privado em prd desde o Sprint 6 |

## 2. Escopo do MVP (MoSCoW)

Crítica ao rascunho que você me deu: login e-mail/Google, SRS simples com 300–500 palavras, painel, metas e deploy cobriam o núcleo, mas **faltavam** itens sem os quais o produto não é seguro de publicar (exclusão/exportação de dados, privacidade, "reportar problema" para item errado, backup) e **sobravam** riscos de prazo (3 recursos que você pediu, TTS, offline e nivelamento, sem decidir onde cortar se estourar). Por isso o MVP é **por incrementos**, com cortes definidos.

| Prioridade | Itens |
|---|---|
| **Must** | Login por e-mail/senha e Google (e o **vínculo por e-mail** que o Keycloak já faz entre eles); catálogo + seed v1 (~500 palavras); sessão de estudo com `simple_v1`, dificuldade e log imutável; exercícios: reconhecimento, digitar, lacuna, imagem→palavra; painel (heatmap, acertos, sequência, previsão, cobertura); metas dinâmicas por ciclo; configurações (idioma, tema, fuso); versão visível; CI/CD tst e prd; backup com restauração testada; LGPD (exportar/excluir) e privacidade; "reportar problema"; créditos/licenças |
| **Should** (você pediu: entram no MVP) | Lembretes por Web Push (Android, desktop, iPhone instalado); nivelamento Sim/Não; TTS (exercício de ouvir); PWA offline com instalar/atualizar |
| **Could** | `fsrs6` + comparação por simulador; frases e expressões além das ~100/40 do MVP; leitura (textos graduados); login GitHub; relações do grafo na interface |
| **Won't (agora)** | Outros assuntos/cursos, tela de administração, XP/ranking, "montar frase", canais de e-mail/Telegram, pilha de observabilidade em prd |

**Walking skeleton (Sprint 1):** repositório com o kit de IA → `api-ci` e `web-ci` verdes (com `gate`) → imagem única → deploy automático em **tst** → login pelo Keycloak → tela "Olá" autenticada com rodapé **"v0.1.0 · API v0.1.0 · tst · a1b2c3d"** → um E2E de fumaça. Se isso roda, o resto é fatia vertical.
**Corte se estourar o prazo:** primeiro o TTS, depois o nivelamento, depois o iPhone no push; offline reduz para "fila de revisões" sem pacote completo. Nunca cortar: backup, LGPD, "reportar problema".

## 3. Roadmap

Total estimado: **11 sprints ≈ 22 semanas (~5 meses)** a 8–10 h/semana, mais folga de 20% `[PREMISSA]`. Pré-requisitos suas, antes do Sprint 0: comprar a VM, registrar `wordloop.com.br`, criar contas (Cloudflare R2, Sonar Cloud, Brevo).

| Sprint | Meta | Entregas | Critério de aceite | Risco principal |
|---|---|---|---|---|
| **0a** (**concluído em 06/10/2026: ver [relatório](../sprint-0a/spike-report.md)**) | Spikes de código | Vite 8 + StyleX + Vitest + Base UI; compatibilidade com **TypeScript 7**; `oapi-codegen` + `hey-api`; `goose` + `sqlc`; **DDL e seed rodando no PostgreSQL**; renomear `main` → `develop` | Cada spike com parecer *aceito/rejeitado* e ADR ajustado; `psql` aplica `content-schema.sql`, `schema-learning.sql` e o seed sem erro | Pilha incompatível (StyleX/TS 7/`hey-api` 0.x) |
| **0b** | Spikes de infraestrutura | Keycloak + BFF + vínculo; Web Push em Android, iPhone instalado e desktop; release-please com PRs de exemplo; deploy de ponta a ponta numa VM; restauração de backup; Sonar com `develop` principal | Push chega nos 3 aparelhos; tag e imagem promovidas; dump restaurado; Sonar analisa uma PR | iOS/Safari; RAM do Keycloak |
| **1** | Esqueleto ponta a ponta | Ver seção 2 | PR → CI verde → tst atualizado → versão visível no rodapé; rollback por redeploy testado | Configuração de Actions e SSH |
| **2** | Conteúdo e catálogo | Migrações de conteúdo; seed v0 (~100 palavras); endpoints do catálogo; tela Palavras; mídia servida; créditos; teste de estilo com OpenMoji | Seed idempotente e conferido; todo item com fonte e licença; você aprova (ou não) o estilo OpenMoji | Estilo infantil; revisão de tradução |
| **3** | Núcleo do SRS | `internal/srs` com `simple_v1`, golden vectors, testes de propriedade, simulador; módulo `learning` (cartões, plano, envio idempotente de revisões, `review_log`) | Testes de propriedade e de integração verdes; replay = estado incremental; reenvio do mesmo UUID não muda nada | Calibragem dos multiplicadores |
| **4** | Estudo na tela | Casca da sessão, plano "quanto tempo hoje?", 4 exercícios, dificuldade em 1 toque, resumo, fila offline | Sessão completa online e offline; Playwright por exercício; alvos ≥ 48 px | UX do teclado móvel; fila offline |
| **5** | Progresso e metas | `study_day`, painel, heatmap, sequência, ciclo de metas com retrospectiva e 3 sugestões, configurações | Painel reflete a sessão na hora; virada de dia por fuso testada | Agregação consistente |
| **6** | Pronto para produção (beta) | Login Google + vínculo (E2E contra o Keycloak), SMTP (Brevo), LGPD (exportar/excluir), privacidade, **ambiente prd**, backups com restauração, release `0.x` em prd (você usando) | Você estuda em prd; backup restaurado em banco descartável; exclusão apaga tudo | Domínio/TLS; e-mail em DKIM |
| **7** | Lembretes | Assinatura, slots, worker, outbox, desafio com deep link, preferências e anti-fadiga, fluxo de instalar no iPhone | Notificação chega, abre na pergunta e registra `source=notification`; 404/410 expira | iOS; entrega do push |
| **8** | Nivelamento, voz e PWA | Teste Sim/Não, exercício de ouvir (TTS), instalar/atualizar/offline completos, seed v1 (~500 revisadas), auditoria de acessibilidade | Nivelamento ≤ 3 min; sem voz inglesa o exercício some; `axe` sem críticos | Vozes variam por aparelho |
| **9** | Endurecimento e **MVP 1.0.0** | Correções do teste de usabilidade, "reportar problema", Lighthouse, documentação, `release-as 1.0.0` | Metas de UX da Fase 5 (1ª resposta ≤ 3 min); sem bug crítico aberto; tag e deploy aprovados | Escopo; revisão de conteúdo |
| **Depois** | `fsrs6` + simulador; frases/expressões; grafo; leitura graduada; GitHub | — | — | — |

## 4. Backlog priorizado do MVP

Tamanho: **S** ≈ 2–4 h · **M** ≈ 6–10 h · **L** ≈ 12–20 h `[PREMISSA]`.

| ID | História | Critério de aceite | MoSCoW | Sprint | Tam. |
|---|---|---|---|---|---|
| F-01 | Repositório com kit de IA e estrutura | `npm run check` roda; hook `guard.mjs` ativo; `develop` padrão | Must | 1 | M |
| F-02 | `api-ci` com `gate` | Lint, testes, `govulncheck`, Sonar; check obrigatório = `gate` | Must | 1 | L |
| F-03 | `web-ci` com `gate` | Biome, ESLint (StyleX), `tsc`, i18n, Vitest, Playwright, Lighthouse | Must | 1 | L |
| F-04 | Deploy automático em tst | Push na `develop` atualiza tst; smoke test passa | Must | 1 | L |
| F-05 | Versão visível | Rodapé e `GET /api/v1/version` mostram web, API, ambiente e commit | Must | 1 | S |
| F-06 | Fluxo de release | release-please abre PR; tag; imagem re-etiquetada; deploy a prd com aprovação | Must | 1/6 | M |
| I-01 | Login e-mail/senha | Keycloak + BFF; cookie `httpOnly`; sair encerra sessão | Must | 1 | L |
| I-02 | Login Google | Entrar por Google cria `app_user` | Must | 6 | M |
| I-03 | Vínculo por e-mail | Senha→Google com mesmo e-mail pede confirmação (e-mail ou senha) e passa a entrar na mesma conta; teste E2E | Must | 6 | M |
| I-04 | Recuperar senha/verificar e-mail | Mensagens chegam por SMTP; link expira | Must | 6 | S |
| I-05 | LGPD: exportar e excluir | JSON completo; exclusão remove dados e usuário no Keycloak | Must | 6 | M |
| I-06 | Privacidade e termos | Páginas publicadas e linkadas | Must | 6 | S |
| C-01 | Migrações de conteúdo | `content-schema.sql` aplicado por `goose` em tst/prd | Must | 2 | S |
| C-02 | Seed v0 → v1 | v0 ~100 palavras (S2); v1 ~500 revisadas (S8); conferência sem violação | Must | 2/8 | L |
| C-03 | API do catálogo | Cursos, unidades, item com textos e mídia por `locale` | Must | 2 | M |
| C-04 | Mídia e créditos | Imagens servidas pelo proxy; tela "Sobre" lista fontes e licenças | Must | 2 | M |
| C-05 | Reportar problema | Botão no item grava relato; lista para você | Must | 9 | M |
| L-01 | `simple_v1` + testes | Golden, propriedades, determinismo | Must | 3 | L |
| L-02 | Simulador de alunos virtuais | Compara retenção e carga entre estratégias | Must | 3 | M |
| L-03 | Cartões e plano da sessão | `POST /study/plan` devolve pacote; limite de novos por minutos | Must | 3 | L |
| L-04 | Envio idempotente de revisões | Lote com UUID; servidor refaz correção e nota; `review_log` só-INSERT | Must | 3 | L |
| L-05 | Geradores e corretores de exercício | Distratores sem irmãos semânticos; erro de digitação tolerado | Must | 3/4 | M |
| U-01 | Casca da sessão e plano | "Quanto tempo hoje?"; sair salva | Must | 4 | M |
| U-02 | 4 exercícios | Reconhecimento, digitar, lacuna, imagem→palavra, com teclado móvel configurado | Must | 4 | L |
| U-03 | Dificuldade em 1 toque | Sugestão por tempo; Enter aceita; no erro "Como foi?" | Must | 4 | M |
| U-04 | Resumo da sessão | Mostra ganhos, não só acerto | Must | 4 | S |
| U-05 | Fila offline | Respostas offline enviadas ao voltar; sem duplicar | Should | 4/8 | L |
| P-01 | Agregação diária | `study_day` na mesma transação da revisão | Must | 5 | M |
| P-02 | Painel | Heatmap, acertos 7 d, sequência, previsão, mais erradas, cobertura | Must | 5 | L |
| P-03 | Metas por ciclo | Pergunta no início; retrospectiva no fim; "manter a mesma" | Must | 5 | L |
| W-01 | Configurações | Idioma, tema (sistema por padrão), fuso, avançado | Must | 5 | M |
| R-01 | Assinatura e permissão | Pedido só depois da 1ª sessão e por toque | Should | 7 | M |
| R-02 | Agendador, slots e outbox | Respeita janela, dias, máximo, silêncio; idempotente | Should | 7 | L |
| R-03 | Desafio por notificação | Deep link abre a pergunta; resposta registrada | Should | 7 | M |
| R-04 | Fluxo iPhone | Instruções de instalar; permissão só depois | Should | 7 | M |
| N-01 | Nivelamento Sim/Não | ≤ 3 min; marcadas entram com S₀ = 3 d | Should | 8 | L |
| W-02 | Ouvir (TTS) | Detecta voz em inglês; sem voz, some | Should | 8 | M |
| W-03 | Instalar/atualizar PWA | `beforeinstallprompt`; aviso de nova versão | Should | 8 | M |
| W-04 | Acessibilidade | `axe` sem críticos; TalkBack e VoiceOver no teste | Must | 8 | M |
| O-01 | Backup e restauração | Dump criptografado diário no R2; restauração testada | Must | 6 | M |
| O-02 | Ambiente prd endurecido | `ufw`, SSH por chave, `fail2ban`, limites de memória | Must | 6 | M |
| O-03 | Segurança no CI | CodeQL, gitleaks, Trivy, ações fixadas por SHA, Renovate | Must | 1/2 | M |

**Definition of Ready:** história com critério de aceite, documento vivo criado e definições confirmadas. **Definition of Done:** `npm run check` verde, contrato e código gerado em dia, textos em pt-BR e en-US, documento vivo atualizado, migração expand/contract, sem segredos no diff, PR com template e squash.

## 5. Lista de decisões em aberto

| # | Item | Quem/quando | Impacto se a resposta mudar |
|---|---|---|---|
| 1 | Horas por semana reais (assumi 8–10) | você, já | muda o calendário (11 sprints) |
| 2 | Comprar a VM e registrar o domínio `wordloop.com.br`; criar contas R2, Sonar e Brevo | você, antes do Sprint 0b | bloqueia spikes de infraestrutura |
| 3 | "Barra de 4 itens e Ler só liberado" é o que você quis? | você | pequena mudança de UI |
| 4 | Documento sobre MIT | você, opcional | só ajusta a seção de licenças |
| 5 | Resultado dos spikes: TypeScript 7, StyleX + Vite 8 + Vitest 5, `hey-api` 0.x, `oapi-codegen` 3.0, RAM do Keycloak, iOS push, atribuição do release-please, CSP × StyleX, licença do Lucide | Sprint 0 | podem trocar uma peça (planos B nos ADRs) |
| 6 | Verificar licenças: Wiktionary, OpenWordNet-PT, StoryWeaver (por livro), Simple English Wikipedia; compatibilidade Tatoeba CC BY 2.0 FR → CC BY-SA; direitos de texto gerado por IA | Sprint 2 | muda quais fontes entram |
| 7 | Redação de privacidade e termos (LGPD) | Sprint 6; considerar revisão jurídica | requisito para abrir a terceiros |
| 8 | Calibrar com dados: S_max = 180 d, multiplicadores do `simple_v1`, regra "Easy só em digitar", novos/dia ≈ minutos/5 | depois de ~1 mês de uso próprio | ajustes numéricos |
| 9 | Ranking real de frequência (NGSL) e `coverage_share` nos itens | Sprint 2 (carga) | base da cobertura e do painel |
| 10 | Marca "Anki": evitar no nome do produto; manter só no nome do repositório | você | só o nome público |

## 6. O planejamento está pronto? (critério da seção 11 do pedido)

| Critério | Situação |
|---|---|
| 1. SRS com fórmulas, exemplos e testes | **Sim.** Fórmulas e exemplo de 6 respostas em `fase-1/01-srs.md`, com script de referência rodável; plano de testes (golden, propriedade, simulação). **Os testes ainda não existem**: nascem no Sprint 3 |
| 2. Dá para criar o repositório e iniciar o Sprint 1 sem decisão grande pendente | **Quase.** Falta o Sprint 0, que **valida** peças ainda não executadas (TS 7, StyleX com Vite 8, PostgreSQL com os DDLs, Keycloak/BFF, push no iPhone) e as suas compras (VM, domínio). Cada risco tem plano B nos ADRs |
| 3. Toda decisão estrutural tem ADR com alternativas | **Sim:** 26 ADRs |
| 4. Pipelines, versões e ambientes ponta a ponta | **Definidos, não executados.** Esqueletos de workflows e deploy com YAML validado por parser; execução real é o Sprint 1 |
| 5. Nenhuma referência inventada | **Sim, com ressalva.** Auditoria em `referencias.md`: cada item tem nível de verificação; um DOI de memória estava errado (Nakata) e foi corrigido; uma referência do arquivo do usuário **não foi encontrada** (Costa/UFRN) e não é citada |

**O que este planejamento não prova:** que a pilha escolhida funciona junta (StyleX 0.19 + Vite 8 + TS 7), que o SQL roda no PostgreSQL, que o Keycloak cabe na VM, que o push chega no iPhone e que o público-alvo acha o app agradável. Tudo isso entra no Sprint 0 e nos testes do Sprint 1.

## 7. Resumo de continuidade (≤ 300 palavras)

Planejamento do **wordloop** concluído em 6 fases; falta o "ok" final e 4 itens de decisão (seção 5: horas/semana, compras de VM/domínio/contas, barra de 4 itens, documento do MIT). Estrutura de `docs/`: `fase-0` a `fase-6`, `adr/` (26 ADRs, gerados por `gen_adrs.py`), `referencias.md` (auditoria com níveis A/B/C/N), `00b-vm-dominio.md`. Roadmap: 11 sprints de 2 semanas (~22 semanas a 8–10 h/semana): **0a/0b** spikes (TS 7, StyleX+Vite 8+Vitest, `hey-api`, Postgres com os DDLs, Keycloak+BFF, push em 3 aparelhos, release-please, deploy, backup, Sonar); **1** esqueleto ponta a ponta com versão visível; **2** conteúdo/catálogo/seed v0; **3** `srs` + `learning`; **4** estudo na tela e fila offline; **5** painel e metas; **6** produção beta (Google, LGPD, backup, prd); **7** lembretes; **8** nivelamento, TTS, PWA, seed ~500 revisadas, a11y; **9** MVP `1.0.0`. MVP = Must + (lembretes, nivelamento, TTS, offline) por incrementos; cortes: TTS, depois nivelamento, depois push no iPhone; nunca backup, LGPD, "reportar problema". Seed: pipeline `JSON → gen_sql.py → SQL idempotente + conferência`; amostra com 31 palavras, 7 expressões, 3 *phrasal verbs*, 10 frases, 8 relações, 10 imagens, todas `draft`; validada em SQLite (idempotente, determinística, teste negativo), **não em PostgreSQL**. Riscos: pilha nova (StyleX 0.x, TS 7, hey-api 0.x), iOS push, RAM do Keycloak, revisão de tradução (você não valida o inglês), licenças de conteúdo. Primeiro passo prático: comprar VM e domínio, criar contas, renomear `main` → `develop` e executar o Sprint 0a.
