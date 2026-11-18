# Progresso e metas (Sprint 5)

> Arquivo: `docs/features/2026-11-18-progresso-metas.md` · Issue: #__ · Branch: `feat/progresso-metas` · Estado: BUILD (PRs 1, 2 e 3a mescladas: #48, #50 e #51; PR 3b pronta; PR 4 pendente)

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
- **PR 3 — P-03, metas (dividida em API e telas; esta PR é a API):** migração `goal_cycle`; `GET /goals/current`, `GET /goals/suggestions`, `POST /goals`; ciclo semanal no fuso da pessoa; 3 sugestões (leve, constante, intensa), personalizar, **manter a mesma**; retrospectiva ao fim do ciclo (dias, minutos, ganhos); não responder **não bloqueia** o estudo (vale a meta anterior). Sem meta, o Início mostra o cartão "Defina sua meta da semana".
- **PR 4 — W-01, configurações:** `GET/PUT /me/settings` (idioma, tema, fuso; fuso validado como nome IANA); tela de configurações com detecção do fuso do aparelho. A virada de dia segue o fuso salvo.

## Decisões
| Tema | Decisão | Alternativa descartada / motivo |
|---|---|---|
| Quem escreve o `study_day` | `learning`, na transação da revisão (P-01 pede a mesma transação) | `progress` escrever por evento: perde a consistência na hora |
| Quem lê | `progress` (só leitura), sem importar `learning` | painel dentro de `learning`: mistura estudo e relatório |
| Cobertura | só "palavras conhecidas" até haver ranking NGSL | % com o seed de ~100 palavras: número enganoso |
| 1ª meta | cartão no Início | tela cheia: bloqueia o caminho para estudar |
| Ciclo de meta | 7 dias a partir de quando a pessoa o abre (dia de estudo de hoje), e não a semana do calendário | semana de segunda a domingo: quem abre na quarta teria um ciclo de 4 dias e uma meta de 5 dias impossível |
| Abrir outra meta | fecha o ciclo aberto (cumprido ou não) e abre o novo numa transação; abrir duas vezes no mesmo dia substitui (um ciclo por dia de início); "manter a mesma" = reenviar os mesmos valores | endpoint próprio de "manter": mais um contrato para o mesmo efeito |
| Fim do ciclo | `ended = true` e o progresso vira a retrospectiva (dias, minutos, palavras novas); a meta **continua valendo** até abrir outra | bloquear o estudo até responder: o UX pede o contrário |
| Sugestões | com ≥ 3 dias de estudo nos últimos 28 dias: 0,8x, 1x e percentil 75 (posto mais próximo) dos minutos por dia estudado, de 5 em 5 min (5 a 240), e dias por semana = estudados/4 (leve −1, intensa +1); sem histórico: 5 min × 3, 10 × 5 e 20 × 7 | média em vez de mediana: um dia longo distorce |
| Telas de meta | feature `goals`: cartão no Início (sem meta, em andamento ou ciclo acabado) e uma só tela `/goals` com a retrospectiva (se acabou), a meta atual e a escolha; sem rota separada `/goals/review` | duas rotas para a mesma conversa: mais navegação sem ganho |
| Dias da semana na meta | a API aceita, a tela **não** pede (valem todos): o filtro só serve aos lembretes (Sprint 7) | formulário com 7 chips agora: complexidade sem uso |
| `study_day.mastered_items` | **fora** da tabela (o esquema de rascunho previa a coluna) | "dominado" ainda não tem definição fechada; entra com a necessidade (YAGNI) |
| Lapso no `study_day` | o cartão perdeu forca naquela resposta (`lapses` do cartão subiu); no preenchimento retroativo, nota Again em cartão já em revisão | contar todo erro: erro em cartão novo não é esquecimento |
| Palavras conhecidas | foto em `study_day.known_lexemes` a cada resposta; o painel lê a visão `user_lexeme_known` na hora | só a foto: ficaria velha até a próxima resposta |
| Onde mora o painel | Início (`/`): a rota junta `HomePage` (auth, atalhos) e `ProgressPanel` (progress) por `children`; as features não se importam | painel em rota própria: esconde o progresso; `progress` importar `auth`: viola a regra de features |
| Heatmap | 12 semanas, colunas de semana (segunda a domingo), tom pela quantidade de respostas (1, 5, 15 e 30) via opacidade do roxo; para leitor de tela é uma imagem com resumo ("estudou em N dias") | 84 células focáveis: ruído |
| Dia de hoje | calculado na API com o fuso e o corte da pessoa (`progress/domain.Today`, cópia da regra do `learning`: módulos não se importam) | ler `local_date` do log: não existe hoje antes da 1ª resposta |

## Checklist de progresso
- [x] Roadmap atualizado (Sprint 4 concluído) e requisitos confirmados
- [x] PR 1: `study_day` + API do painel (migração 00015, soma na transação, módulo `progress`, contrato)
- [x] PR 2: painel na web (feature `progress`, Início, i18n, Vitest e Playwright)
- [x] PR 3a: metas, API (migração 00016, ciclo de 7 dias, sugestões, retrospectiva)
- [x] PR 3b: metas, telas (cartão no Início, escolher/manter a meta, retrospectiva)
- [ ] PR 4: configurações
- [ ] `npm run check` verde em cada PR
- [ ] PRs abertas (com aprovação do usuário)

## Verificação
- **PR 1:** `npm run check` verde (API: gofmt, vet, golangci-lint, estrutura, arch-lint com os componentes novos de `progress`, `go test -short`; web: tudo). Testes unitários cobrem a regra de dia/fuso (madrugada, Tóquio, Lisboa no horário de verão), a sequência (hoje ainda não estudou, buraco, dia sem resposta), a taxa de acertos, a previsão, o painel de quem nunca estudou, o contrato HTTP (200, 400, 401) e a soma única por resposta no `study_day` (reenvio não soma; falha em `BumpDay` desfaz o log).
  - **Não rodei** os testes de integração com Postgres (sem Docker aqui): compilam com `-tags integration` e o CI roda. Cobrem a migração 00015 (inclusive o preenchimento retroativo, executando o SQL da própria migração), as consultas do painel (previsão no fuso da pessoa, atrasadas em hoje, suspensas fora, palavras conhecidas, mais erradas sem erro de digitação, isolamento entre pessoas) e a soma no mesmo ciclo da revisão.
  - Os dois endpoints ainda não têm tela: o painel na web é a PR 2.
- **PR 2:** `npm run check` verde. Vitest cobre a grade do heatmap (alinhamento na segunda, 84 dias, tons), a altura das barras e o painel (tudo preenchido, sem estudo hoje, sem acertos em 7 dias, quem nunca estudou, falha do servidor sem derrubar o Início, atalhos mantidos). Playwright (2 cenários) rodou localmente no Edge e passa, e a captura foi conferida a olho: cartões em coluna única, sem rolagem horizontal.
  - Pendente: o cartão "Defina sua meta da semana" e o progresso da meta entram na PR 3; cobertura em % espera o NGSL.
- **PR 3a (API de metas):** `npm run check` verde (golangci-lint, arch-lint, estrutura, `go test -short`, web). Testes unitários cobrem os limites da meta (0/241 minutos, 0/8 dias, dia da semana repetido ou inválido, menos dias permitidos que a meta), o progresso (só os dias do ciclo, minutos arredondados, dias que faltam, "cumprida"), o fim do ciclo (retrospectiva e meta mantida), a substituição no mesmo dia, o fechamento do ciclo antigo, as sugestões (valores iniciais, histórico, limites, ordem leve ≤ constante ≤ intensa, 28 dias) e o contrato HTTP (200, 201, 400 e 401, inclusive sem o cabeçalho de proteção).
  - **Não rodei** os testes de integração com Postgres (sem Docker aqui): compilam com `-tags integration`; o CI roda. Cobrem a migração 00016, o fechamento transacional dos ciclos antigos, a substituição no mesmo dia, o isolamento entre pessoas e os `CHECK` do banco.
  - Esta PR não tem tela: o cartão "Defina sua meta da semana" e a retrospectiva são a PR 3b.
- **PR 3b (telas de meta):** `npm run check` verde. Vitest cobre as escolhas e a validação do "personalizar" (0 e 241 minutos, 0 e 8 dias, decimal, texto, limites), a porcentagem da barra, e a tela inteira: cartão sem meta, em andamento, no último dia, com ciclo acabado e com falha do servidor (o Início segue); escolher sugestão (corpo e cabeçalho de proteção), "Constante" já escolhida, personalizar válido e inválido (botão desligado), retrospectiva cumprida e não cumprida, "Manter a mesma" com os dias da semana, falha ao salvar. Playwright (2 cenários) rodou localmente no Edge: do cartão à escolha e de volta ao Início com a meta; alvos de toque ≥ 48 px e sem rolagem horizontal. Capturas conferidas a olho.
  - Pendente: lembretes que usam os dias da semana (Sprint 7) e a cobertura em % (NGSL).
