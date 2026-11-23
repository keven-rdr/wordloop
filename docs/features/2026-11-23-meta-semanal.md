# Meta da semana define o tempo de estudo

> Arquivo: `docs/features/2026-11-23-meta-semanal.md` · Issue: — · Branch: `feat/meta-semanal` · Estado: PR

## Contexto
A tela Estudar perguntava "quanto tempo hoje?" toda vez (5, 15, 30 ou 60 minutos) e a meta (ciclo de 7 dias a partir do
dia em que foi aberta) não tinha ligação com esse tempo. O dono quer escolher a meta uma vez por semana: o tempo
escolhido vale para todos os dias de segunda a domingo, e na segunda seguinte o app pergunta de novo.

Fase B do plano (A: estudo que avança e varia, #60; C: gamificação; D: layout desktop, #59; E: IA via MCP).

## Perguntas e respostas dos requisitos
| # | Pergunta | Recomendação | Resposta |
|---|---|---|---|
| 1 | O tempo vale todos os dias ou só em dias escolhidos? | Todos os dias | "o mesmo tempo para todos os dias da semana, seg a dom" |
| 2 | Quando perguntar de novo? | Na segunda | "chegou na outra segunda vai perguntar de novo" |

## Definições confirmadas
- A meta é o tempo por dia, o mesmo de segunda a domingo.
- Estudar não pergunta mais o tempo.
- Na segunda seguinte (a semana acabou) o app pede a meta da nova semana antes de estudar.

## Decisões
- **Ciclo = semana ISO** no dia de estudo da pessoa (fuso e corte): `domain.WeekStart` acha a segunda; aberta no meio
  da semana, a meta vale até domingo. `days_target` fica opcional (ausente = 7) e `weekdays` é ignorado (todos os dias).
  As sugestões continuam vindo do histórico (0,8×, 1× e p75 dos minutos), agora todas com 7 dias.
- **Transição**: um ciclo antigo de 7 dias corridos pode ter começado depois da segunda da meta nova. Abrir a meta
  fecha **todos** os outros ciclos abertos (não só os anteriores) e a meta atual é sempre a aberta. Ciclos antigos que
  ninguém trocou acabam no próprio fim (até 6 dias depois desta versão).
- **O tempo da sessão vem da web, não da API**: Estudar lê `GET /goals/current` (cliente gerado) e manda os minutos da
  meta no `POST /study/plan` (5 a 240; meta abaixo de 5 vira 5). Descartado ler a meta dentro do `learning`: exigiria
  uma porta para o módulo `progress` só para isso, e a API continua aceitando qualquer tempo de 5 a 240.
- **Sem meta da semana** (nunca houve ou o domingo passou): Estudar mostra "Escolher a meta da semana", que abre
  `/goals?from=study`; ao salvar, volta para Estudar. **Sem conseguir ler a meta** (sem rede): estuda 5 minutos e avisa,
  em vez de travar.
- A escolha da meta pede só "Minutos por dia"; "Manter a mesma" repete o tempo.

## Checklist de progresso
- [x] Contrato: `openapi.yaml` (descrições de `/goals*`, `GoalRequest.days_target` opcional, `weekdays` obsoleto) e `npm run gen`
- [x] API: domínio da meta (semana), consultas (`CurrentGoal`, `CloseOlderGoals`), testes de unidade e de integração
- [x] Web: Estudar com a meta, escolha só de minutos, volta para Estudar, textos `pt-BR`/`en-US`, Vitest e Playwright
- [x] Documentação: este arquivo
- [ ] `npm run check` verde (ver Verificação)
- [ ] PR aberta

## Verificação
- `go test ./internal/progress/...`: verde. O teste de integração novo (ciclo antigo de quarta substituído pela meta
  da semana) compila com `-tags integration`; roda no CI (sem Docker na máquina local).
- `npx vitest run src/features/goals`: 30 testes verdes.
- `npx playwright test e2e/goals e2e/study`: 12 verdes, incluindo "na segunda, Estudar pede a meta antes e volta com o
  tempo dela".
