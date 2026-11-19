# ADR 0031 — Fontes do conteúdo v1: NGSL (vocabulário) e Tatoeba (frases e traduções)

**Status:** aceita (2026-11-18), com pendência de revisão jurídica. Complementa o [ADR 0024](0024-licencas-codigo-mit-e-conteudo-cc-by-sa-4-0.md).

## Contexto
O seed v0 tinha 105 palavras e 10 frases, todas em rascunho, escritas à mão. Para o app servir de verdade são precisos milhares de itens, com fonte e licença conhecidas e crédito mostrado a quem usa. Ninguém revisa traduções à mão neste projeto (uma pessoa, sem orçamento).

## Decisão
- **Vocabulário:** os 2.100 lemas mais frequentes da **NGSL 1.2** (New General Service List, CC BY-SA 4.0), em ordem de frequência, divididos em lições de 30 posições. O nível CEFR é aproximado pela faixa de frequência (A1 até 600, A2 até 1.500, B1 depois). A NGSL não traz classe gramatical nem tradução: `pos` e a glosa em pt-BR são escritas no projeto, **assistidas por IA e sem revisão humana**, e o crédito diz isso.
- **Frases e exemplos:** pares inglês↔português do **Tatoeba** (CC BY 2.0 FR), com tradução humana. Filtros automáticos: 3 a 9 palavras, uma frase só, sem números ou símbolos, só palavras do vocabulário (formas flexionadas da NGSL, `Tom` e `Mary`), lista de termos bloqueados, descarte de marcas de português de Portugal, traduções que inventam ou perdem um nome. Cada frase entra na lição da **palavra mais rara que usa**.
- **Crédito:** cada item guarda o número das duas frases (`eng#id/por#id`) e o nome dos autores (`attribution`); a tela de créditos mostra as fontes e a licença de cada uma.
- **Reprodutível:** as fontes são baixadas por `content/tools/fetch_sources.py` (url, data e sha256 em `content/seed/v1/sources.json`) e **não** vão para o Git; os dados escolhidos (`content/seed/v1/*.json`) vão. Migrações geradas, em lote.
- **Publicação:** tudo nasce `published` depois da validação automática (`content/tools/validate.py`); sem revisão humana antes.

## Consequências
Volume (4.100 itens novos) sem custo recorrente e com origem rastreável por frase. A qualidade das glosas e das traduções depende de IA e de colaboradores do Tatoeba: erros chegam aos usuários. O subconjunto CC0 do Tatoeba quase não tem português (17 frases), por isso não serve.

## Pendências (não são aconselhamento jurídico)
- Confirmar com revisão jurídica que material CC BY 2.0 FR pode compor um conjunto distribuído sob CC BY-SA 4.0 com o crédito que fazemos (a prática comum é que sim, com atribuição).
- Direitos sobre texto gerado por IA continuam em aberto (já marcados no [seed v0](../fase-6/seed/README.md)).

## Alternativas descartadas
Escrever tudo à mão (volume pequeno por PR); Wiktionary/Simple Wikipedia (ruído de marcação e licença por trecho); IA para frases (sem origem rastreável e sem tradução humana); só o subconjunto CC0 do Tatoeba (quase sem português).

## Fonte
`docs/features/2026-11-18-conteudo-e-montar-frases.md`, `content/NOTICE.md`.
