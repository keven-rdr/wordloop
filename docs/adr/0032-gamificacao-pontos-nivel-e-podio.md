# ADR 0032 — Gamificação: pontos, nível e pódio com apelido opcional

**Status:** aceita (2026-11-25). Substitui em parte o [ADR 0026](0026-experiencia-sem-xp-ranking-nem-mascote-sequencia.md): passa a haver pontos, nível e ranking; continuam valendo a sequência "sem culpa", os sons desligados por padrão e a ausência de mascote.

## Contexto
O ADR 0026 recusou XP e ligas e deixou anotado o risco: "pode ter menos apelo que apps com pontos: validar no teste de usabilidade". Nos testes, o dono pediu gamificação: meta da semana, pontos que o app lembra por pessoa e um pódio entre todos os usuários logados.

## Decisão
- **Pontos derivados do estudo**, sem livro-razão separado: cada resposta soma no `study_day.points`, na mesma transação do log (acerto no reconhecimento 10, na produção 15, erro 2; multiplicador pelo nível do item: A1 ×1, A2 ×1,2, B1 ×1,5, B2 ou mais ×2), com teto de 3.000 por dia. Na leitura entram os bônus da meta: +50 por dia com o tempo da meta cumprido e +300 pela semana com os 7 dias cumpridos. O `review_log` continua sendo a verdade (os pontos de dias antigos foram recalculados a partir dele).
- **Nível** pela curva do total de pontos: o 1º nível pede 100 pontos e cada um seguinte pede 100 + nível × 50 a mais.
- **Pódio** com duas abas, **Semana** (de segunda a domingo, a mesma da meta) e **Geral** (desde sempre), com empate dividindo a posição (`RANK`).
- **Privacidade:** ninguém aparece para os outros sem querer. Para entrar no ranking a pessoa escolhe um **apelido** (3 a 20 letras, números ou `_`, único) e liga a participação; o nome do perfil nunca aparece. Quem não participa vê a própria posição só para si. O perfil do ranking entra na exportação de dados e sai junto com a conta.
- Módulo novo `leaderboard`, só leitura sobre `study_day` e `goal_cycle` (como o `progress`), com a tabela própria `leaderboard_profile`.

## Consequências
Mais motivo para voltar todo dia e uma comparação entre pessoas. Risco de quem estuda pelos pontos e não pela palavra: os pontos só nascem de respostas a pacotes montados pelo servidor e corrigidas por ele, e o teto diário limita o abuso. A barra inferior passa a ter 5 itens (Ranking).

## Alternativas descartadas
Livro-razão de pontos com regras configuráveis (mais tabelas e telas de admin sem necessidade agora); ligas com promoção e rebaixamento (complexidade); ranking com o nome do perfil para todos (expõe quem não quer).

## Fonte
`docs/features/2026-11-25-gamificacao.md`.
