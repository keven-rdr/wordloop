# Núcleo do SRS e módulo de aprendizado (Sprint 3)

> Arquivo: `docs/features/2026-10-19-nucleo-srs.md` · Issue: #__ · Branch: `feat/srs-learning` · Estado: VERIFY

## Contexto
O catálogo está pronto (Sprint 2). Falta o que transforma palavras em aprendizado: o **agendador** (quando revisar cada cartão), o **estado do usuário** no banco e a **API de estudo**. O Sprint 3 do [roadmap](../fase-6/06-roadmap-backlog.md) entrega `internal/srs` (biblioteca pura) com `simple_v1`, golden vectors, testes de propriedade e simulador; e o módulo `learning` (cartões, plano da sessão, envio idempotente de revisões, `review_log` só-INSERT). A especificação está em [fase-1/01-srs.md](../fase-1/01-srs.md) e o esquema rascunhado em [fase-3/schema-learning.sql](../fase-3/schema-learning.sql) e [openapi-sketch.yaml](../fase-3/openapi-sketch.yaml). A tela de estudo é o Sprint 4; geradores de exercício (L-05) ficam para o 3/4.

## Perguntas e respostas dos requisitos
| # | Pergunta | Recomendação | Resposta |
|---|---|---|---|
| 1 | Como dividir o sprint em PRs? | 3 PRs em ordem: `srs` → simulador → `learning` | Recomendação aceita (19/10/2026) |
| 2 | "Foi distração" / erro de digitação vale Hard sem lapso? | Sim (≤ 1 letra, palavras ≥ 5 letras) | Recomendação aceita (19/10/2026) |
| 3 | "Dominado" = S ≥ 90 dias sem lapso nas 2 últimas revisões? | Sim | Recomendação aceita (19/10/2026) |
| 4 | Easy só vale em exercício de produção? | Sim; em múltipla escolha conta como Good | Recomendação aceita (19/10/2026) |
| 5 | Palavras novas por dia no início | minutos ÷ 5, teto inicial 8 (A1), zero novos se a fila passar de 2× a meta | Recomendação aceita (19/10/2026) |

## Definições confirmadas
- **PR 1 — `internal/srs`:** `Scheduler` (`Name`, `Next` pura, `Retrievability`), `simple_v1` (1º acerto 1/2/4 d; acerto `S' = (S + 0,5·atraso)·m` com m = 1,2/2,0/3,0; erro `max(1, 0,25·S)`, lapso +1; S ≤ 365; fuzz determinístico por `hash(card_id, n)`), curva `R(t,S) = (1 + F·t/S)^−0,1542`, domínio `ln(1+S)/ln(181)`, mapeamento de nota (errou → Again; Difícil/Médio/Fácil → Hard/Good/Easy; dica/explicação/imagem antes limitam a Hard), `replay(log) → CardState`. Testes: golden vectors conferidos com `docs/fase-1/srs_reference.py`, propriedades, determinismo, replay = incremental.
- **PR 2 — simulador:** alunos virtuais com esquecimento próprio (verdade ≠ modelo); mede retenção real × alvo, revisões/dia, itens com R ≥ 0,9 no dia 30/90 e calibração; roda como teste/comando, não em produção.
- **PR 3 — `learning`:** migrações aditivas (`user_settings`, `device`, `card`, `study_session`, `review_log` + gatilho append-only, `review_void`); `POST /study/plan` (pacote de cartões por minutos: 5/15/30/60; novos = minutos ÷ 5 com teto inicial 8; revisões por menor R; irmãos receptivo × produtivo); `POST /study/reviews` em lote com UUID do cliente (reenvio não muda nada; o servidor refaz correção, nota e estado); dia lógico vira às 04:00 no fuso do usuário; erros em Problem Details.
- Estratégia padrão `simple_v1`; `fsrs6` fica para depois, mesma interface.

## Decisões
| Tema | Decisão | Alternativa descartada / motivo |
|---|---|---|
| Biblioteca | `internal/srs` sem I/O e sem dependências de módulos | agendador dentro de `learning`: não testável isolado |
| Estado | derivado do `review_log` (replay); `card` guarda o estado atual como cache | só o estado: impede trocar de estratégia |
| Migrações | aditivas, uma por funcionalidade (`00012…`) | uma migração grande: difícil de revisar |
| Correção | o servidor refaz a correção e a nota | confiar no cliente: nota viesada e adulterável |

## Checklist de progresso
- [x] PR 1: `internal/srs` com `simple_v1`, golden vectors, propriedades, determinismo e replay (cobertura 100%)
- [x] PR 2: simulador de alunos virtuais e relatório de calibração (`go run ./cmd/srs-sim`)
- [x] PR 3: contrato (`openapi.yaml` + `npm run gen`), migrações 00012/00013, SQL (`sqlc`), casos de uso, handlers, testes de unidade e integração (o de integração roda no CI)
- [x] Documentação: este arquivo
- [x] `npm run check` verde em cada PR
- [ ] PRs abertas (com aprovação do usuário)

## Verificação
- PR 1 (#35): `internal/srs` com cobertura de 100%, `check:all` verde.
- PR 2: `internal/srs/sim` (cobertura de 100%) e `cmd/srs-sim`. Cenário de referência (90 dias, 8 novas/dia, até 60 revisões/dia, semente 1), `simple_v1` × escada fixa (1-3-7-14-30-60-120 d):

| aluno | estratégia | retenção | lapsos | sabidas d90 |
|---|---|---|---|---|
| cuidadoso | simple_v1 / escada | 0,943 / 0,944 | 139 / 169 | 652 / 647 |
| mediano | simple_v1 / escada | 0,888 / 0,869 | 294 / 320 | 465 / 365 |
| irregular | simple_v1 / escada | 0,740 / 0,688 | 477 / 505 | 45 / 48 |

- Leitura: o `simple_v1` retém mais que a escada fixa quando o aluno estuda menos ou erra a autoavaliação; com aluno cuidadoso empatam. **Calibração:** com R previsto de 0,9–1,0 o observado foi 0,945 (cuidadoso), 0,906 (mediano) e 0,861 (irregular); em 0,8–0,9 o aluno irregular acertou só 0,732. Ou seja, o modelo é bem calibrado para quem estuda direito e **otimista para quem estuda pouco**: é o risco "calibragem dos multiplicadores" do roadmap, a revisar com dados reais (decisão 8). Os alunos virtuais usam uma curva de esquecimento exponencial própria (verdade ≠ modelo).
- PR 3 (`learning`): `POST /study/plan` e `POST /study/reviews`; migrações `00012` (`user_settings`, `card`, `study_session`) e `00013` (`review_log` com gatilho append-only e `review_void`). Verificado: `npm run check:all` verde; testes de unidade e HTTP do módulo (cobertura 94–98%); migrações, gatilho, `ON CONFLICT` (reenvio) e as consultas de seleção executados no PGlite. **Não rodei o teste de integração no Postgres real** (sem Docker): roda no CI.
- Decisões tomadas na construção: tabela `device` adiada (o `device_id` é só uma coluna, sem FK); estabilidade e métricas do log em `double precision` (sem deriva entre o estado e o replay); resposta digitada em branco conta como erro ("não sei") e fica no log; resposta fora de ordem reproduz o log inteiro (`srs.Replay`); cartões receptivos criados no plano e reaproveitados enquanto não respondidos; produtivo nasce com S ≥ 7 d e revisão há ≥ 1 dia; distratores sem ligação semântica e sem a mesma tradução.
