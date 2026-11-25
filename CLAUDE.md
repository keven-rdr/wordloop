@AGENTS.md

<!-- Só o que é específico do Claude Code vai aqui. Regras de projeto ficam no AGENTS.md (uma fonte só). -->

## Claude Code

- Hooks de bloqueio em `.claude/settings.json` (`.claude/hooks/guard.mjs`): `.env*`, código gerado, arquivo > 300 linhas, string de UI crua, cor crua. Se um hook negar, **corrija a causa**; não tente contornar.
- Antes de concluir uma tarefa: rode `npm run check` e, em mudanças relevantes, `/code-review`.
- Push de branch de trabalho: use `/wl-push` — sai com o link da PR já preenchido (título Conventional Commits e descrição no template).
- Subpastas têm `AGENTS.md` próprio (`api/`, `web/`): são lidos quando você abre arquivos lá.
