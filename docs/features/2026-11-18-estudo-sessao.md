# Estudo na tela (Sprint 4)

> Arquivo: `docs/features/2026-11-18-estudo-sessao.md` · Issue: #__ · Branch: `feat/estudo-sessao` · Estado: BUILD (3 PRs prontas, a aguardar publicação)

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
- **PR 2 — L-05:** exercícios **lacuna** (`cloze`) e **imagem → palavra** (`image`). Ambos são variantes do cartão **produtivo** (S ≥ 7 d) e se respondem **digitando**; a correção é a mesma do digitar (tolera 1 erro em palavras ≥ 5 letras). O **servidor** escolhe a variante, de forma determinística por cartão (`hash(card_id, reps)`): unidade "Palavrinhas" (`little-words`) com exemplo que contém o termo → sempre lacuna; nos demais, 1ª fatia → lacuna se couber (senão imagem), 2ª fatia → imagem se houver e S < 21 d (*fading*), 3ª → PT→digitar. Lacuna mostra a frase com `____` e a tradução da frase como dica; imagem→palavra mostra só a imagem (sem a tradução). A imagem **é a pergunta**: não ativa `image_before_answer` (Fácil continua possível). Contrato: `exercise` ganha `cloze` e `image` (cartão e resposta) e o cartão ganha `cloze_text`. Migração aditiva alarga o `CHECK` de `review_log.exercise_type`.
- **PR 3 — U-05:** fila offline em IndexedDB com o UUID do cliente; faixa com contador de respostas pendentes; reenvio ao voltar a conexão, sem duplicar; a sessão continua com o pacote local. Detalhes confirmados na construção:
  - **Cada resposta vai para o aparelho na hora** (antes de qualquer envio). O pacote da sessão também é guardado ao chegar do servidor; **recarregar a página no meio da sessão retoma de onde parou** (fila = pacote menos os cartões já acertados).
  - **Envio:** a cada 10 respostas, ao terminar/sair, ao voltar a conexão (`online`), ao abrir o plano e a cada 30 s enquanto sobrar algo. Lotes de até 200, sempre com os mesmos UUIDs; um envio por vez.
  - **Terminar sem rede não trava:** o resumo sai na hora, **provisório** para as respostas sem confirmação (usa o palpite do app) e avisa quantas esperam. Servidor confirmou → conta o que o servidor corrigiu; rejeitada → não conta.
  - **Faixa** "Sem conexão. Suas respostas ficam salvas e são enviadas depois." e contador "N respostas aguardando envio" no plano e na sessão; some quando tudo foi enviado.
  - **Limpeza:** sessão encerrada ou pacote vencido (`expires_at`) saem do aparelho; resposta pendente **nunca** é apagada. Terminada a sessão, voltar para a URL dela leva ao plano (não dá para repetir).
- Corrigir resposta, nota e agendamento **só no servidor**; o cliente só mostra feedback provisório (`client_correct`).

## Decisões
| Tema | Decisão | Alternativa descartada / motivo |
|---|---|---|
| Estado da sessão | Zustand local (fila de cartões, índice, respostas pendentes); servidor só pelo TanStack Query | estado do servidor em store: duplica a fonte da verdade |
| Envio | em lote ao sair, a cada N respostas e ao concluir; UUID por resposta | uma chamada por resposta: ruim offline e para a bateria |
| Menu | editar `BottomNav` (já é arquivo central de Palavras/Início) | só botão no Início: esconde a função principal |
| Limiares de dificuldade | constantes isoladas, calibráveis com dados reais | sem sugestão: foge do critério U-03 |
| Fila offline | IndexedDB direto (≈100 linhas, sem biblioteca), com a memória da página como alternativa quando ele não existe; contrato único de armazenamento testado nos dois | `idb`/Dexie: dependência a mais para 2 tabelas |
| Resumo sem rede | provisório (palpite do app) + aviso de pendentes | bloquear até o servidor responder: trava o fim da sessão sem rede |

## Checklist de progresso
- [x] Roadmap e doc do Sprint 3 fechados
- [x] PR 1: contrato já pronto (`npm run gen`), feature `study` (plano, sessão, escolher, digitar, dificuldade, resumo), rotas, menu, i18n, testes (Vitest + Playwright)
- [x] PR 2: lacuna e imagem → palavra (contrato, migração 00014, variantes e correção no servidor, telas)
- [x] PR 3: fila offline (IndexedDB, retomada ao recarregar, faixa, reenvio sem duplicar)
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
  - Pendente: imagem na apresentação do cartão novo e TTS (Sprint 8) e a faixa offline (PR 3).
- **PR 2:** `go test` do `learning` e do `srs` verde (variantes, `Blank`, determinismo, fading da imagem, rejeição de variante que não cabe na direção do cartão); integração Postgres **compila** com a tag `integration`, mas **não rodou** (sem Docker): o CI roda e cobre o `CHECK` alargado pela migração 00014. Vitest cobre lacuna e imagem (frase sem a palavra, tradução como dica, dica limita a nota, queda para o digitar quando faltam frase ou imagem); Playwright ganhou 2 cenários.
  - A variante é do servidor e determinística (`hash(card_id, reps)`); o app só mostra o que veio no plano. A correção aceita qualquer variante digitada em cartão produtivo e recusa `cloze`/`image` em cartão receptivo.
  - `Blank` troca a primeira ocorrência da palavra **inteira** no exemplo (sem letra colada); formas flexionadas ("runs" para "run") não geram lacuna e o cartão cai no digitar.
  - Instalei o Go 1.27.4 nesta máquina (winget) para gerar o código e rodar a API.
- **PR 3:** Vitest cobre a fila (idempotência, rejeitada, limpeza, lotes de 200, envio único, retomada, resumo provisório), o armazenamento (IndexedDB via `fake-indexeddb` e memória, mesmo contrato) e o fluxo na tela (sem rede → resumo provisório → reenvio com os mesmos UUIDs; recarregar retoma; sessão encerrada não repete). Playwright ganhou 1 cenário offline; os 25 specs rodaram localmente no Edge e passam. O CI pegou um erro que o Vitest não pega: sem rede o TanStack Query **pausa** mutações (`networkMode: 'online'`), então o envio e o plano travavam; as duas mutações agora usam `networkMode: 'always'` e o teste unitário simula a pausa com `onlineManager`. Dependência nova só de desenvolvimento: `fake-indexeddb`.
  - **Limites conhecidos:** o reenvio automático só acontece com o app aberto (envio em segundo plano/Service Worker fica para o Sprint 8, PWA); um pedido recusado inteiro pelo servidor (400/401) volta a ser tentado a cada 30 s; o resumo provisório não se atualiza sozinho depois que a conexão volta.
