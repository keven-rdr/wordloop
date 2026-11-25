# Gamificação: pontos, nível e pódio

> Arquivo: `docs/features/2026-11-25-gamificacao.md` · Issue: — · Branch: `feat/gamificacao` · Estado: PR

## Contexto
O dono quer começar a gamificação: meta da semana, pontos que o app lembra por pessoa e um pódio entre todos os
usuários logados. O ADR 0026 recusava XP e ranking; o [ADR 0032](../adr/0032-gamificacao-pontos-nivel-e-podio.md) o
revisa. Fase C do plano (A: estudo que varia, #60; B: meta semanal, #61; D: layout desktop, #59; E: IA via MCP).

## Perguntas e respostas dos requisitos
| # | Pergunta | Recomendação | Resposta |
|---|---|---|---|
| 1 | Período do pódio? | Semana + geral | Semana + geral |
| 2 | Como identificar as pessoas (privacidade)? | Apelido com participação opcional | Apelido, opt-in |

## Definições confirmadas
- Pódio com abas Semana (de segunda a domingo) e Geral.
- Só aparece quem escolhe um apelido e liga a participação; o nome do perfil nunca aparece.
- Os pontos ficam guardados por pessoa (derivados do estudo).

## Decisões
- **Pontos por resposta**, somados no `study_day.points` na mesma transação do log (`learning/domain.ReviewPoints`):
  acerto 10 no reconhecimento e 15 na produção, vezes o nível do item (A1 1; A2 1,2; B1 1,5; B2+ 2); erro 2; teto de
  3.000 por dia. A migração `00025` recalcula os dias antigos a partir do `review_log`.
- **Bônus da meta** na leitura, pela visão `user_day_score`: +50 por dia com o tempo da meta cumprido e +300 pela
  semana de meta com os 7 dias cumpridos.
- **Nível** pelo total: do nível n para o n+1 são 100 + n × 50 pontos (2 aos 100, 3 aos 250, 4 aos 450).
- **Posição** conta só quem participa (`RANK`, empate divide a posição). Quem não participa vê a própria linha, com a
  posição que teria, sem empurrar ninguém.
- **Apelido**: 3 a 20 letras sem acento, números ou `_`, único sem diferenciar maiúsculas, com uma lista curta de
  trechos bloqueados. Apelido em uso responde 409; inválido, 400.
- Módulo novo `api/internal/leaderboard` (só leitura sobre `study_day`/`goal_cycle` + a tabela `leaderboard_profile`),
  endpoints `GET /leaderboard`, `GET /leaderboard/me`, `GET`/`PUT /leaderboard/profile`. O perfil entra na exportação
  de dados (`ranking`) e sai junto com a conta (`ON DELETE CASCADE`); os pontos do dia também vão na exportação.
- Web: feature `leaderboard` com a rota `/ranking` (pódio em degraus, lista, minha linha, participação) e o cartão de
  pontos no Início; o item "Ranking" entra na barra pelo `feature.ts` (agora com 5 itens). Cores das medalhas como
  tokens (`gold`, `silver`, `bronze`, `onMedal`).
- Fica para depois: aviso de pontos no resumo da sessão.

## Checklist de progresso
- [x] Contrato: `openapi.yaml` (tag `leaderboard`, 4 operações, 4 esquemas) e `npm run gen`
- [x] API: migração 00025 (expand), SQL (`sqlc`), pontos no `learning`, módulo `leaderboard`, testes (unidade, HTTP e integração)
- [x] Web: rota, feature, hooks gerados, textos `pt-BR`/`en-US`, tokens StyleX, Vitest e Playwright
- [x] Documentação: este arquivo e o ADR 0032
- [ ] `npm run check` verde
- [ ] PR aberta

## Verificação
- `go test ./internal/leaderboard/... ./internal/learning/...`: verde. Os testes de integração (ranking, bônus da meta,
  apelido único, pontos no `study_day`, exportação) compilam com `-tags integration` e rodam no CI (sem Docker local).
- `npx vitest run src/features/leaderboard`: 5 verdes. `npx playwright test e2e/leaderboard`: verde (pódio, minha
  linha, alvos de 48 px, sem rolagem horizontal no celular).
