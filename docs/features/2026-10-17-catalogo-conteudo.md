# Catálogo e conteúdo (Sprint 2)

> Arquivo: `docs/features/2026-10-17-catalogo-conteudo.md` · Issue: #__ · Branch: `feat/catalogo-conteudo` · Estado: BUILD

## Contexto
O login está completo e ativo em tst. Falta o que faz do wordloop um app de vocabulário: **o conteúdo**. O Sprint 2 do [roadmap](../fase-6/06-roadmap-backlog.md) entrega o esquema de conteúdo no banco, um seed v0 (~100 palavras), a API do catálogo (C-03), a tela **Palavras**, mídia servida e créditos. Não depende de nenhuma conta externa: tudo roda em tst.

Ponto de partida já existente (rascunho, nunca executado em PostgreSQL): [`docs/fase-2/content-schema.sql`](../fase-2/content-schema.sql) e a amostra em [`docs/fase-6/seed`](../fase-6/seed/README.md) (31 palavras, 7 expressões, 3 *phrasal verbs*, 10 frases, 8 relações, 10 imagens; validada só em SQLite).

## Perguntas e respostas dos requisitos
| # | Pergunta | Recomendação | Resposta |
|---|---|---|---|
| 1 | Quais ~100 palavras entram no seed v0? | Estender a amostra de 31 palavras com as mais frequentes do NGSL (faixa 1), só substantivos, verbos e adjetivos concretos, com textos originais do projeto (CC BY-SA 4.0), todos `draft` | Recomendação aceita (17/10/2026) |
| 2 | Quem enxerga itens `draft` em tst? | A tela mostra `published` para todos e `draft` só ao admin, com selo "rascunho" | Recomendação aceita (17/10/2026) |
| 3 | Onde ficam as imagens (OpenMoji)? | Arquivos SVG estáticos servidos pelo web (`/media/...`), com só o manifesto no banco; créditos gerados do `attribution` | Recomendação aceita (17/10/2026) |
| 4 | O que a tela Palavras faz no Sprint 2? | Lista paginada (`page`, `size`, `sort`), busca por texto, filtro por unidade e CEFR, detalhe do item com exemplos e crédito | Recomendação aceita (17/10/2026) |
| 5 | Idioma de ajuda | `pt-BR` e `en-US` por `item_text.locale`; a UI escolhe pelo idioma do usuário | Recomendação aceita (17/10/2026) |

## Definições confirmadas
- Seed v0: as palavras mais frequentes do NGSL (faixa 1), textos originais do projeto, tudo `draft`.
- Itens `draft`: só o admin vê, com selo; os demais veem só `published`.
- Imagens: SVG estático servido pelo web; no banco só manifesto e crédito.
- Tela Palavras: lista paginada, busca por texto, filtro por unidade e CEFR, detalhe com exemplos e crédito.
- Idioma de ajuda: `pt-BR` e `en-US` por `item_text.locale`, pelo idioma do usuário.

## Decisões
| Tema | Decisão | Alternativa descartada / motivo |
|---|---|---|
| Migrações | Esquema de conteúdo em migrações aditivas novas (`00008…`), divididas por funcionalidade, com `goose` e `sqlc` | uma migração gigante: difícil de revisar e de reverter |
| Seed | `JSON → gen_sql.py → SQL idempotente + conferência` já existente; sem sobrescrever decisão humana (`status`, `content_version`) | seed manual: sem repetição nem conferência |
| Validação | Primeira execução real em PostgreSQL (Testcontainers/PGlite), não só SQLite | confiar no SQLite: não provou a sintaxe |

## Checklist de progresso
- [ ] Contrato: `openapi.yaml` (+ `npm run gen`)
- [ ] API: migrações, SQL (`sqlc`), casos de uso, handlers, testes (unidade + integração)
- [ ] Seed v0 gerado e conferido (toda linha com fonte e licença)
- [ ] Web: rota Palavras, feature, hooks gerados, textos `pt-BR`/`en-US`, tokens StyleX, testes (Vitest + Playwright)
- [ ] Mídia servida e créditos
- [ ] Documentação: este arquivo e `docs/`
- [ ] `npm run check` verde
- [ ] PR aberta (com aprovação do usuário)

## Verificação
A preencher.
