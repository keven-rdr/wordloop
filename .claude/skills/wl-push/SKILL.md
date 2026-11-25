---
name: wl-push
description: Push da branch do wordloop com a PR já preparada - título Conventional Commits e descrição no template do repositório, entregues num link que abre o formulário do GitHub preenchido. Use SEMPRE que for fazer push de uma branch de trabalho (feat/, fix/, chore/...) ou quando pedirem para abrir/preparar a PR.
argument-hint: "[título da PR, opcional]"
allowed-tools: Bash(git status *) Bash(git log *) Bash(git diff *) Bash(git push *) Bash(git rev-parse *) Bash(node scripts/pr-link.mjs *) Bash(npm run check)
---

# /wl-push $ARGUMENTS

Todo push de branch de trabalho sai com a PR pronta: o usuário só clica no link, confere e cria.

## Passos
1. **Pré-condições.** Branch de trabalho (nunca `develop`/`main`), árvore limpa (`git status`), commits Conventional. Se ainda não rodou nesta sessão depois da última mudança: `npm run check` e guarde o resumo (linhas `ok`/`FALHOU` e a última linha).
2. **Push.** `git push -u origin <branch>` (nunca `--force` nem `--no-verify`). Se já estiver em dia, diga isso e siga.
3. **Título.** Conventional Commits, em português, no máximo 100 caracteres depois do `: ` — o merge é por squash e o título vira o commit na `develop` (o check `pr-title` reprova fora do padrão). Escopo pelo módulo ou feature: `feat(goals): ...`, `fix(study): ...`. Use `$ARGUMENTS` se vier.
4. **Descrição.** Escreva em `.git/wl-pr/<branch com / trocado por ->.md` (dentro do `.git`: nunca entra em commit) seguindo `.github/pull_request_template.md` **sem os comentários `<!-- -->`**:
   - **Contexto e objetivo**: o problema e por quê; link do documento vivo `docs/features/...`.
   - **O que foi feito**: frases no nível do comportamento, sem lista de arquivos.
   - **Decisões relevantes**: o que um revisor questionaria e a alternativa descartada.
   - **Como testar**: passos reproduzíveis e o resultado esperado.
   - **Definition of Done**: marque `[x]` só o que foi verificado; o que não foi, deixe `[ ]` e diga por quê numa linha.
   - **Saída do check** no `<details>`: o resumo real do passo 1, nunca inventado.
   - Última linha: `🤖 Generated with [Claude Code](https://claude.com/claude-code)`.
5. **Link.** `node scripts/pr-link.mjs --title "<título>"` (lê a descrição do passo 4). O script valida o título, recusa descrição com comentário do template e avisa se o push não está em dia. Se avisar que o link ficou longo demais, entregue o link curto e diga que a descrição está no arquivo para colar.
6. **Entrega.** Mostre o link (base `develop`) e o título. Não crie a PR por API nem com o token do Git: quem clica em "Create pull request" é o usuário.

## Regras
- Push é saída para terceiros só na branch de trabalho; publicar a PR continua sendo decisão do usuário.
- Se o `check` falhou, diga qual etapa e por quê na descrição e na resposta; não marque o item do DoD.
- Mais commits na mesma branch depois de a PR existir: só o push (passo 2); se a descrição mudou de sentido, regere o arquivo e diga o que atualizar na PR.
