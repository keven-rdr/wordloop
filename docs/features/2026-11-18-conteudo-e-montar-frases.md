# Conteúdo em massa liberado para todos + montar frases

> Arquivo: `docs/features/2026-11-18-conteudo-e-montar-frases.md` · Issue: #__ · Branch: `feat/conteudo-e-montar-frases` · Estado: BUILD

## Contexto
O catálogo tinha 105 palavras, 7 expressões, 3 phrasal verbs e 10 frases, e **tudo em `draft`**: catálogo e fila de estudo só devolvem itens `published` a quem não é admin, e **nenhuma migração, comando ou tela publicava itens**. Um usuário comum não via nem estudava nada. O dono pediu: (1) muito mais palavras e frases, (2) liberar para todos, (3) o sistema também **montar sentenças**.

## Perguntas e respostas dos requisitos
| # | Pergunta | Recomendação | Resposta |
|---|---|---|---|
| 1 | De onde vêm as palavras e frases? | Listas abertas + tradução assistida por IA | **Listas abertas + tradução assistida** (18/11/2026) |
| 2 | Tamanho do primeiro lote | ~2.000 palavras + ~2.000 frases | **~2.000 + ~2.000** |
| 3 | Como liberar com traduções sem revisão humana? | Publicar tudo após validação automática | **Publicar tudo após validação automática** |
| 4 | O que é "montar sentenças"? | Montar a frase em inglês com peças, a partir da frase em pt-BR | **Peças em inglês a partir do pt-BR** (toque e teclado) |

## Definições confirmadas
- **Conteúdo (Parte A):** vocabulário = 2.100 lemas mais frequentes da NGSL (menos os 103 que o v0 já tem) → **2.000 palavras novas**; **2.100 frases** do Tatoeba (par inglês↔português); **70 lições** de 30 posições de frequência. Tudo `published`, inclusive os 125 itens do v0 (migração própria).
- **Montar frases (Partes B e C):** novo exercício `assemble` para itens `sentence` em cartão produtivo; peças embaralhadas das palavras da frase; correção exata (sem tolerância a erro de digitação), sem diferenciar maiúsculas e pontuação.

## Decisões
| Tema | Decisão | Alternativa descartada / motivo |
|---|---|---|
| Fontes | NGSL (CC BY-SA 4.0) e Tatoeba (CC BY 2.0 FR); [ADR 0031](../adr/0031-fontes-do-conteudo-v1.md) | escrever tudo à mão (volume); subconjunto CC0 do Tatoeba (17 frases em português) |
| Glosas e POS | escritas no projeto, assistidas por IA, **sem revisão humana**, avisado nos créditos | pedir revisão a terceiros: sem orçamento |
| Frase ↔ lição | cada frase vai para a lição da **palavra mais rara** que usa (≥ 100% das palavras no vocabulário, com formas flexionadas) | lição por tema: exige classificar 2.100 frases |
| Exemplo de cada palavra | frase curta do Tatoeba só com palavras tão frequentes quanto ela; prefere a forma exata (para o cloze) | exemplo de IA: sem origem rastreável |
| Qualidade das frases | 3 a 9 palavras, uma frase só, sem número/símbolo, lista de termos bloqueados, descarte de português de Portugal, tradução que inventa ou perde `Tom`/`Mary` | aceitar tudo: o corpus tem muito ruído |
| Carga no banco | migrações em lote (`INSERT` de 500 linhas): ~50 instruções em vez de ~25 mil (1 por linha custa ~0,15 s no Neon; o v0 já estourou 120 s) | uma migração por linha: 1 h+ de deploy |
| Publicação | itens novos nascem `published`; o v0 vira `published` por migração própria (`00023`) | exigir um comando de publicar: ninguém o rodaria |
| Ordem de estudo | por lição; dentro da lição, palavras antes de frases (`NewItems`) | ordem alfabética misturada |
| Títulos das lições | uma chave i18n `units.lesson` ("Lição A1 · 3") em vez de 70 chaves | uma tradução por lição |
| Frases sem lição do v0 (7 expressões, 3 phrasal verbs, 10 frases) | continuam sem unidade: entram depois de todas as lições | atribuí-las a lições: fora de escopo |

## Checklist de progresso
- [x] Pipeline: `content/tools` (fetch, build, validate, gen) e 2.000 glosas em `content/seed/v1/glosses`
- [x] Dados: `content/seed/v1/{units,words,sentences}.json` e migrações `00020`–`00023`
- [x] Testes das ferramentas e das migrações geradas (atualizadas, integridade referencial, lote)
- [x] API: ordem de estudo (palavras antes de frases); testes de integração com a carga real
- [x] Web: título das lições (`themeTitle`) e contagem em "itens"
- [x] CI: `content-ci.yml` e etapa `content` no `npm run check`; ADR 0031 e `content/NOTICE.md`
- [ ] Parte B: API de montar frase (contrato, migração `00019`, domínio, testes)
- [ ] Parte C: web de montar frase (componente, teclado, i18n, testes)
- [ ] `npm run check` verde
- [ ] PR aberta (com aprovação do usuário)

## Verificação
- `python content/tools/validate.py`: 70 lições, 2.000 palavras, 2.100 frases. `python -m unittest discover -s content/tools`: 17 testes (filtros, escolha, validação, migrações). Mutação feita de propósito (trocar um `unit_id`): os testes de migração falham.
- **Não rodei** os testes de integração com Postgres (sem Docker aqui): compilam com `-tags integration`; o CI roda. Eles conferem a carga real (2.105 palavras, 7 expressões, 3 phrasal verbs, 2.110 frases; nada fora de `published`; todo item com termo e tradução; 76 unidades, todas com itens para o aluno; 4 fontes nos créditos) e a ordem (lição 7 começa pelas palavras). Os testes antigos de visibilidade voltam tudo a rascunho antes de testar a regra.
- Pendências do dono: revisão jurídica das licenças (Tatoeba CC BY 2.0 FR em obra CC BY-SA 4.0; texto de IA), qualidade das glosas, contato do responsável nos textos de privacidade.
