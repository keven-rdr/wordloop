# Estudo na tela (Sprint 4)

> Arquivo: `docs/features/2026-11-18-estudo-sessao.md` · Issue: #__ · Branch: `feat/estudo-sessao` · Estado: BUILD

## Contexto
O núcleo do SRS está pronto (Sprint 3): `POST /study/plan` entrega o pacote de cartões e `POST /study/reviews` recebe respostas em lote, de forma idempotente. Falta a **tela de estudo**. O Sprint 4 do [roadmap](../fase-6/06-roadmap-backlog.md) pede: casca da sessão, plano "quanto tempo hoje?", 4 exercícios, dificuldade em 1 toque, resumo e fila offline. Critério: sessão completa online e offline; Playwright por exercício; alvos ≥ 48 px. UX em [fase-5/05-ux-ui-fluxos.md](../fase-5/05-ux-ui-fluxos.md) (seções 4.2, 5 e 6). Itens: U-01, U-02, U-03, U-04, U-05 e L-05 (parte 4).

## Perguntas e respostas dos requisitos
| # | Pergunta | Recomendação | Resposta |
|---|---|---|---|
| 1 | Como dividir o sprint em PRs? | 3 PRs: sessão online (escolher + digitar) → lacuna e imagem→palavra (L-05) → fila offline | Recomendação aceita (18/11/2026) |
| 2 | Como entrar em `/study` pelo menu? | Editar o `BottomNav` (item "Estudar") e botão no Início; registrar a exceção à regra 7 | Recomendação aceita (18/11/2026) |
| 3 | Limiares da dificuldade sugerida no acerto | ≤ 3 s Fácil, ≤ 8 s Médio, acima Difícil; no digitar soma 4 s; constantes em um arquivo | Recomendação aceita (18/11/2026) |

## Definições confirmadas
- **PR 1 — sessão online:** rotas `/study` (plano: 5/15/30/60 min), `/study/session/:id` (loop) e `/study/summary/:id`; hooks gerados `planSession` e `submitReviews`; exercícios **escolher** e **digitar** (já suportados pela API); estados `pergunta → verificando → acerto | erro → dificuldade → próximo`; dificuldade em 1 toque (acerto: Fácil/Médio/Difícil com sugestão por tempo; erro: "Não sabia / Quase lembrei / Foi distração"); erro devolve o cartão à fila em 3 cartões; sair (✕) pede confirmação e **envia o que foi respondido**; resumo mostra ganhos; atalhos de teclado (`Enter`, `1–3`, `A–D`, `Esc`); alvos ≥ 48 px; feedback em `aria-live="polite"`.
- **PR 2 — L-05:** exercícios **lacuna** e **imagem → palavra** (contrato: novos valores de `exercise` e campos do cartão; geradores e corretores no servidor, sem irmãos semânticos nos distratores).
- **PR 3 — U-05:** fila offline em IndexedDB com o UUID do cliente; faixa com contador de respostas pendentes; reenvio ao voltar a conexão, sem duplicar; a sessão continua com o pacote local.
- Corrigir resposta, nota e agendamento **só no servidor**; o cliente só mostra feedback provisório (`client_correct`).

## Decisões
| Tema | Decisão | Alternativa descartada / motivo |
|---|---|---|
| Estado da sessão | Zustand local (fila de cartões, índice, respostas pendentes); servidor só pelo TanStack Query | estado do servidor em store: duplica a fonte da verdade |
| Envio | em lote ao sair, a cada N respostas e ao concluir; UUID por resposta | uma chamada por resposta: ruim offline e para a bateria |
| Menu | editar `BottomNav` (já é arquivo central de Palavras/Início) | só botão no Início: esconde a função principal |
| Limiares de dificuldade | constantes isoladas, calibráveis com dados reais | sem sugestão: foge do critério U-03 |

## Checklist de progresso
- [x] Roadmap e doc do Sprint 3 fechados
- [x] PR 1: contrato já pronto (`npm run gen`), feature `study` (plano, sessão, escolher, digitar, dificuldade, resumo), rotas, menu, i18n, testes (Vitest + Playwright)
- [ ] PR 2: lacuna e imagem → palavra (contrato, API, tela)
- [ ] PR 3: fila offline
- [ ] Documentação: este arquivo
- [ ] `npm run check` verde em cada PR
- [ ] PRs abertas (com aprovação do usuário)

## Verificação
- **PR 1:** `npm run check` verde (Biome, ESLint, `tsc`, i18n, estrutura, Vitest e regras nas linhas alteradas). Vitest cobre as regras de resposta, o reducer, os atalhos e a sessão completa (plano → escolher → digitar → resumo, erro, dica, sair, falha ao salvar e reenvio com os mesmos UUIDs). Spec Playwright em `web/e2e/study/` (escolher, digitar, sair; confere alvos ≥ 48 px). **Não rodei o Playwright localmente** (sem os navegadores instalados): roda no CI. Também **não conferi a tela no navegador** com a API real.
- Decisões tomadas na construção (PR 1):
  - Estado da sessão em `useReducer` (puro, testável), não em Zustand: a biblioteca ainda não está no projeto e o reducer basta. O pacote e o resumo ficam no cache do TanStack Query; **recarregar a página no meio da sessão volta ao plano** (a persistência local chega com a fila offline, PR 3).
  - A sessão cobre a tela toda (`position: fixed`), inclusive o menu inferior, sem editar o `AppLayout`.
  - O menu ganhou o item "Estudar" (única edição em arquivo central, já prevista) e o Início ganhou o botão "Estudar agora".
  - Erro de digitação (1 letra, palavras ≥ 5) e acerto com dica não perguntam dificuldade: o servidor aplica Difícil sem lapso. O cartão errado volta daqui a 3 cartões.
  - Atalhos: A–D/1–4 escolhem, Enter verifica/continua, 1–3 respondem a dificuldade (com o foco nos botões, Enter e Espaço ficam com o botão), Esc abre/fecha a confirmação de sair.
  - Pendente para o PR 2/3: imagem na apresentação, TTS e a faixa offline.
