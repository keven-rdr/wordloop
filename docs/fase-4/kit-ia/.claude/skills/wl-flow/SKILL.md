---
name: wl-flow
description: Fluxo enxuto de uma issue do wordloop até a PR, com gates (documento vivo, check, verificação). Use ao implementar uma issue ou feature.
disable-model-invocation: true
argument-hint: "[número da issue]"
allowed-tools: Bash(gh issue view *) Bash(gh pr create *) Bash(gh pr view *) Bash(git status *) Bash(git diff *) Bash(git log *) Bash(git switch *) Bash(git add *) Bash(git commit *) Bash(git push *) Bash(npm run *)
---

# /wl-flow $ARGUMENTS

Uma única máquina de estados. **Só avance quando o gate do estado atual passar.** Estado atual = o primeiro da lista cujo gate ainda não passou; diga-o em uma linha no início de cada resposta.

| # | Estado | Faça | Gate para sair |
|---|---|---|---|
| 1 | **ISSUE** | `gh issue view $ARGUMENTS`; resuma o objetivo em 3 linhas | objetivo entendido; sem ambiguidade (senão pergunte, **uma pergunta por vez**, com sua resposta recomendada) |
| 2 | **DOC** | crie `docs/features/AAAA-MM-DD-slug.md` a partir de `_TEMPLATE.md`: contexto, perguntas e respostas, **definições confirmadas** | o usuário confirmou as definições (sem confirmação não se escreve código) |
| 3 | **BRANCH** | `git switch -c feat/<slug>` a partir de `develop` atualizada | branch criada; árvore limpa |
| 4 | **BUILD** | implemente seguindo `docs/guides/new-feature.md`; marque o checklist do documento vivo | checklist do doc em dia |
| 5 | **VERIFY** | `npm run check`; leia a saída inteira | check **verde**; se falhar, volte a BUILD |
| 6 | **REVIEW** | `/code-review`; corrija o que for relevante | sem achados abertos |
| 7 | **PR** | commits Conventional; `gh pr create --base develop` com o template `.github/pull_request_template.md` (Contexto e objetivo / O que foi feito / Decisões relevantes / Como testar; **sem lista de arquivos**) | descrição preenchida; usuário aprovou publicar |
| 8 | **DONE** | cole o resultado do `check` e o link da PR | só aqui diga "concluído" |

## Regras do fluxo
- Publicar a PR é uma ação visível a terceiros: **peça confirmação** antes do `gh pr create`.
- Nunca declare algo pronto sem a evidência (saída do `check`). Se um check foi pulado, diga qual e por quê.
- Mudança de contrato (OpenAPI) → `npm run gen` e inclua o código gerado no mesmo commit.
- Em dúvida entre duas abordagens: **recomende uma** e mostre a alternativa descartada em 1 linha.
