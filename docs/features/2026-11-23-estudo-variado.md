# Estudo que avança e varia

> Arquivo: `docs/features/2026-11-23-estudo-variado.md` · Issue: — · Branch: `feat/estudo-variado` · Estado: PR

## Contexto
Nos testes do dono, toda sessão começava pelas mesmas palavras, nada era sorteado, palavras marcadas "Fácil" voltavam
logo e 5 minutos traziam pouquíssimas palavras. O estado por pessoa já era salvo (`card`, `review_log`, `study_day`); as
causas eram de seleção e de nota:

- 5 minutos davam **1** palavra nova (`minutos / 5`, teto 8), e os cartões produtivos ocupavam essas vagas primeiro;
  acima de 40 revisões pendentes, nenhuma palavra nova.
- As novas saíam sempre na mesma ordem para todo mundo (lição, nível, ordem alfabética); distratores e a posição da
  resposta certa eram fixos por cartão.
- "Fácil" na múltipla escolha vale "Médio" (regra 4.4 do SRS), então a palavra volta em 2 dias.

É a fase A do plano maior (meta semanal, gamificação, layout desktop e IA via MCP vêm em PRs próprios).

## Perguntas e respostas dos requisitos
| # | Pergunta | Recomendação | Resposta |
|---|---|---|---|
| 1 | "Fácil" no reconhecimento deve valer Easy? | Botão "Já sabia" no 1º contato (sinal explícito) | Botão "Já sabia" |

## Definições confirmadas
- Mais palavras novas por sessão, crescendo com o tempo; o tempo vai de 5 a 240 minutos (vem da meta diária na fase B).
- As novas mudam a cada sessão e há palavras de desafio de lições adiante.
- "Já sabia" só no primeiro contato com a palavra.

## Decisões
- **Palavras novas** = `6 + 0,65 × minutos`, teto 45 (5 min: 9 · 15: 16 · 30: 26 · 60: 45). Na avalanche (pendentes
  acima de 2× o limite) as novas caem **pela metade** em vez de zerar.
- **Produtivos com cota própria** (metade das novas): não tiram a vaga das palavras novas.
- **Semente por sessão** = o id da sessão. Ela sorteia a ordem das novas **dentro da lição** (a ordem das lições e
  "palavras antes das frases" continuam), os lugares das novas no meio das revisões, os distratores e a posição da
  resposta certa. As revisões mantêm a ordem de menor retenção (quem para no meio revisou as mais urgentes). A mesma
  sessão dá sempre a mesma ordem (o pacote pode ser remontado igual).
- **Desafio**: 1 em cada 5 novas vem de lições duas ou mais posições adiante da primeira lição pendente (`StretchItems`),
  só palavras/expressões. Deixa a sessão mais difícil sem pular a ordem das lições. Descartado: calcular o nível CEFR da
  pessoa (mais SQL, pouco ganho agora; a fase E traz desafios gerados por IA).
- **"Já sabia"** (`difficulty: known`): 4º botão (tecla 4) só no 1º contato (cartão novo, escolher, 1ª tentativa na
  sessão, sem dica). Vale Easy com S = 21 dias; o produtivo nasce no dia seguinte e confere. Num cartão já respondido
  vale "Fácil" (para não perder a resposta de uma sessão retomada); no digitar ou com ajuda é rejeitado
  (`invalid_answer`). Descartado: "Fácil" = Easy no 1º contato, porque no simulador a retenção do aluno "mediano" caía de
  0,888 para 0,861 (abaixo da escada fixa).
- Migração `00024_review_known.sql` (expand): só alarga o CHECK de `review_log.difficulty_chosen` com `known`.

## Checklist de progresso
- [x] Contrato: `openapi.yaml` (`PlanRequest.minutes` 5–240, `ReviewInput.difficulty` + `known`) e `npm run gen`
- [x] API: migração (expand), SQL (`NewItems` com semente, `StretchItems`, `Distractors` com semente), casos de uso, testes
- [x] Web: botão "Já sabia" (pt-BR/en-US), tecla 4, testes do reducer e do teclado
- [x] Documentação: este arquivo e `docs/fase-1/01-srs.md` (4.4 e 4.5)
- [ ] `npm run check` verde: tudo verde menos `vitest` (ver Verificação)
- [ ] PR aberta

## Verificação
- `go test ./internal/srs/... ./internal/learning/...`: verde, incluindo o simulador (`TestSimpleV1AgainstTheLadder`).
- Testes de integração (Postgres) atualizados e compilados com `-tags integration`; rodam no CI (sem Docker na máquina local).
- `npm run check` (2026-11-23): API toda verde (gofmt, vet, golangci-lint, estrutura, arquitetura, testes); web verde em
  biome, eslint, tsc, i18n e estrutura. `vitest` falha em 24 testes de componente (`Study`, `Offline`, `Assemble`,
  `Variants`) **também no `develop` sem esta mudança**, na mesma máquina (Windows): falha do ambiente, não regressão.
- Conhecido e aceito: "Já sabia" conta como palavra conhecida na hora (S = 21 ≥ 7 em `user_lexeme_known`); depois de
  recarregar a página o botão pode reaparecer num cartão já tentado, e aí o servidor o trata como "Fácil".
