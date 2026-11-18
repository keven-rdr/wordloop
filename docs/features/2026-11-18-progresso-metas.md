# Progresso e metas (Sprint 5)

> Arquivo: `docs/features/2026-11-18-progresso-metas.md` · Issue: #__ · Branch: `feat/progresso-metas` · Estado: BUILD (PR 1 pronta; PRs 2 a 4 pendentes)

## Contexto
O estudo na tela está pronto (Sprint 4, #43, #44, #46). Falta mostrar **o que a pessoa ganhou** e **quanto ela quer estudar**. O Sprint 5 do [roadmap](../fase-6/06-roadmap-backlog.md) entrega P-01 (agregação diária), P-02 (painel), P-03 (metas por ciclo) e W-01 (configurações). Critério: o painel reflete a sessão na hora; a virada de dia por fuso é testada. Modelo de dados em [schema-learning.sql](../fase-3/schema-learning.sql) (`study_day`, `goal_cycle`, `user_lexeme_known`); contrato esboçado em [openapi-sketch.yaml](../fase-3/openapi-sketch.yaml); fluxo "Fim do ciclo" em [05-ux-ui-fluxos.md](../fase-5/05-ux-ui-fluxos.md) §4.4.

## Perguntas e respostas dos requisitos
| # | Pergunta | Recomendação | Resposta |
|---|---|---|---|
| 1 | Como dividir o sprint em PRs? | 4 PRs: API de progresso → painel na web → metas → configurações | Recomendação aceita (18/11/2026) |
| 2 | O que mostrar em "cobertura", sem o ranking NGSL carregado? | "N palavras conhecidas" agora; o % de cobertura entra com o NGSL (item 9 do backlog) | Recomendação aceita (18/11/2026) |
| 3 | Como definir a 1ª meta (o onboarding só vem no Sprint 8)? | Cartão no Início com 3 sugestões; não bloqueia o estudo | Recomendação aceita (18/11/2026) |
| 4 | Configurações entram no sprint? | Sim: idioma, tema e fuso; "Avançado" (agendador, retenção) fica para o FSRS | Recomendação aceita (18/11/2026) |

## Definições confirmadas
- **PR 1 — P-01 e API do painel (`api/`):**
  - Migração aditiva: tabela `study_day` (como no esquema) e preenchimento retroativo a partir do `review_log` (idempotente).
  - O módulo `learning` atualiza o `study_day` **na mesma transação** da linha do `review_log` (`reviews`, `correct`, `lapses`, `new_items`, `active_ms`); resposta repetida (mesmo UUID) não soma de novo. O dia é o `local_date` já calculado com fuso e hora de corte.
  - Módulo novo `internal/progress` (domain · app · adapters), **só leitura**: `GET /progress/dashboard` (sequência, acertos dos últimos 7 dias, hoje, previsão de 7 dias, palavras conhecidas, mais erradas) e `GET /progress/heatmap?from&to`.
  - **Estudou no dia = pelo menos 1 revisão.** Sequência = dias seguidos com estudo até hoje; se hoje ainda não estudou, ontem não quebra a sequência.
  - Palavras conhecidas = lemas da visão `user_lexeme_known` (receptivo em revisão com S ≥ 7 d).
- **PR 2 — P-02, painel na web:** o Início vira o painel (botão "Estudar agora", sequência, heatmap de ~12 semanas, acertos 7 d, previsão, mais erradas, palavras conhecidas). Feature nova `progress`; sem arquivo central novo.
- **PR 3 — P-03, metas:** migração `goal_cycle`; `GET /goals/current`, `GET /goals/suggestions`, `POST /goals`; ciclo semanal no fuso da pessoa; 3 sugestões (leve, constante, intensa), personalizar, **manter a mesma**; retrospectiva ao fim do ciclo (dias, minutos, ganhos); não responder **não bloqueia** o estudo (vale a meta anterior). Sem meta, o Início mostra o cartão "Defina sua meta da semana".
- **PR 4 — W-01, configurações:** `GET/PUT /me/settings` (idioma, tema, fuso; fuso validado como nome IANA); tela de configurações com detecção do fuso do aparelho. A virada de dia segue o fuso salvo.

## Decisões
| Tema | Decisão | Alternativa descartada / motivo |
|---|---|---|
| Quem escreve o `study_day` | `learning`, na transação da revisão (P-01 pede a mesma transação) | `progress` escrever por evento: perde a consistência na hora |
| Quem lê | `progress` (só leitura), sem importar `learning` | painel dentro de `learning`: mistura estudo e relatório |
| Cobertura | só "palavras conhecidas" até haver ranking NGSL | % com o seed de ~100 palavras: número enganoso |
| 1ª meta | cartão no Início | tela cheia: bloqueia o caminho para estudar |
| `study_day.mastered_items` | **fora** da tabela (o esquema de rascunho previa a coluna) | "dominado" ainda não tem definição fechada; entra com a necessidade (YAGNI) |
| Lapso no `study_day` | o cartão perdeu forca naquela resposta (`lapses` do cartão subiu); no preenchimento retroativo, nota Again em cartão já em revisão | contar todo erro: erro em cartão novo não é esquecimento |
| Palavras conhecidas | foto em `study_day.known_lexemes` a cada resposta; o painel lê a visão `user_lexeme_known` na hora | só a foto: ficaria velha até a próxima resposta |
| Dia de hoje | calculado na API com o fuso e o corte da pessoa (`progress/domain.Today`, cópia da regra do `learning`: módulos não se importam) | ler `local_date` do log: não existe hoje antes da 1ª resposta |

## Checklist de progresso
- [x] Roadmap atualizado (Sprint 4 concluído) e requisitos confirmados
- [x] PR 1: `study_day` + API do painel (migração 00015, soma na transação, módulo `progress`, contrato)
- [ ] PR 2: painel na web
- [ ] PR 3: metas
- [ ] PR 4: configurações
- [ ] `npm run check` verde em cada PR
- [ ] PRs abertas (com aprovação do usuário)

## Verificação
- **PR 1:** `npm run check` verde (API: gofmt, vet, golangci-lint, estrutura, arch-lint com os componentes novos de `progress`, `go test -short`; web: tudo). Testes unitários cobrem a regra de dia/fuso (madrugada, Tóquio, Lisboa no horário de verão), a sequência (hoje ainda não estudou, buraco, dia sem resposta), a taxa de acertos, a previsão, o painel de quem nunca estudou, o contrato HTTP (200, 400, 401) e a soma única por resposta no `study_day` (reenvio não soma; falha em `BumpDay` desfaz o log).
  - **Não rodei** os testes de integração com Postgres (sem Docker aqui): compilam com `-tags integration` e o CI roda. Cobrem a migração 00015 (inclusive o preenchimento retroativo, executando o SQL da própria migração), as consultas do painel (previsão no fuso da pessoa, atrasadas em hoje, suspensas fora, palavras conhecidas, mais erradas sem erro de digitação, isolamento entre pessoas) e a soma no mesmo ciclo da revisão.
  - Os dois endpoints ainda não têm tela: o painel na web é a PR 2.
