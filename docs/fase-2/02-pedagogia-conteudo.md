# Fase 2 — Pedagogia, conteúdo e modelo de dados de conteúdo

> Data: 2026-10-06. Produto: **wordloop**. Rótulos: `[PREMISSA]` assumi; `[VERIFICAR]` não confirmei em fonte primária; `[DECISÃO SUA]` só você decide.
> **Como li as fontes:** DOIs conferidos no Crossref (título/autor/ano). "Resumo lido" = li o resumo ou página oficial na busca de hoje; o resto é conteúdo de memória ou da sua descrição. Esquema SQL completo: [content-schema.sql](content-schema.sql).
> Decisões fechadas na Fase 1: dominado = S ≥ 90 d; *Foi distração* = Hard sem lapso; Easy só em exercício de digitar.

## 1. Tabela de decisões

| Decisão | Recomendação | Alternativas descartadas | Por quê |
|---|---|---|---|
| Unidade de ensino | **1 item = 1 sentido, 1 pergunta, 1 resposta curta** | uma ficha por palavra com todos os sentidos | "Cartão atômico" evita interferência; *bank* = banco/margem vira 2 itens, o 1º só no nível certo |
| Métrica de progresso | **Cobertura de vocabulário por texto**: % dos tokens de um texto cujo lema você já conhece | contagem bruta de palavras; só nível CEFR | É o que prediz compreensão (Schmitt et al. 2011: relação praticamente linear) e é o critério para liberar leitura |
| Primeira imagem | **OpenMoji** (CC BY-SA 4.0) para conceitos e abstratos; Wikimedia Commons selecionado para concretos | Pexels/Pixabay/Unsplash; IA generativa | Estilo único, licença clara, redistribuível. Pexels/Pixabay vetam redistribuição "como arquivo avulso"/competir; Unsplash exige *hotlink* |
| Onde ficam as imagens | **Fora do repositório** (volume + manifesto `media` com hash e licença) | versionar no Git | Repositório público: evita redistribuir arquivos com termos próprios e deixa o Git leve |
| Conexões entre palavras | Grafo tipado, liberado **só quando os dois itens têm S ≥ 21 d** | ensinar conjuntos semânticos juntos | Interferência (seção 6) |
| Nivelamento | **Teste Sim/Não com pseudopalavras** por faixa de frequência, curto e adaptativo | LexTALE; teste de múltipla escolha longo | LexTALE foi feito para avançados; Sim/Não é rápido e funciona em nível baixo |
| Textos de leitura | **Misto**: StoryWeaver (CC BY) e Simple English Wikipedia para curadoria + geração offline com vocabulário controlado e validação automática | só curadoria; só IA em tempo real | Nível A1 quase não tem texto aberto; geração em lote é grátis no uso e verificável por script |
| Licença do conteúdo | Diretório `content/` em **CC BY-SA 4.0** com `NOTICE` | tudo MIT | NGSL e OpenMoji são Share-Alike; ver seção 7 |

## 2. Itens e sentidos

| Tipo (`kind`) | Exemplo | Pergunta típica | Observação |
|---|---|---|---|
| `word` | *sort* (verbo, sentido 1) | EN→PT, lacuna, ouvir | por sentido, com `lexeme` (lema + classe) separado |
| `expression` | *by the way* | lacuna, PT→EN | colocações vêm de listas ou do Tatoeba |
| `phrasal_verb` | *give up* | lacuna em frase | uma ficha por **sentido** (Garnier & Schmitt 2015, ver seção 7) |
| `sentence` | *I live in a hot country.* | lacuna, montar frase | alimenta o desafio da notificação |
| `dialogue` / `paragraph` / `text` | mini-diálogo, texto graduado | compreensão e lacunas | só com cobertura suficiente (seção 3) |

Regras de qualidade do item `[PREMISSA]`: (1) um sentido por item; (2) resposta curta (≤ 4 palavras); (3) toda palavra de nível A1 tem **1 frase de exemplo curta** com vocabulário já ensinado; (4) até **3 sentidos** por palavra nos 1.000 primeiros, e o sentido 2 só entra depois do 1 em review; (5) tradução PT-BR revisada por você (nativo).

## 3. Trilha até a leitura

Seu ponto de partida é A1 (`[PREMISSA]`). Os degraus abaixo são meus; os números de cobertura que cito têm fonte, os de cada faixa serão **calculados no seed** com a frequência do NGSL.

| Degrau | Conteúdo | Meta de cobertura | Libera |
|---|---|---|---|
| T0 | ~300 palavras mais frequentes + ~40 frases fixas (cumprimentos, pedidos) | — | exercícios de ouvir e lacuna |
| T1 | até ~1.000 palavras + ~100 frases curtas | cobertura do **texto graduado** ≥ 95% | primeiros textos de 3–5 frases |
| T2 | ~1.000–2.000 + ~100 colocações/*phrasal verbs* + frases compostas | ≥ 95% em textos de 50–100 palavras | parágrafos |
| T3 | NGSL completo (**2.809 palavras ≈ 92% de textos gerais**, ≈ 95% em muitas séries de TV, segundo o [site oficial do NGSL](https://www.newgeneralservicelist.com/new-general-service-list)) | ≥ 95% (guiada) / ≥ 98% (livre) | **página inteira** |
| T4 | listas irmãs (falada, acadêmica) e conexões | 98% | leitura livre de textos autênticos graduados |

**Por que 95% e 98%:** [Laufer & Ravenhorst-Kalovski 2010](https://doi.org/10.64152/10125/66648) (resumo lido): limiar mínimo ~4.000–5.000 famílias para 95%, ótimo ~8.000 para 98%. [Nation 2006](https://doi.org/10.3138/cmlr.63.1.59) (resumo lido): 8.000–9.000 famílias para 98% em texto escrito (BNC); 6.000–7.000 para fala. [Schmitt, Jiang & Grabe 2011](https://doi.org/10.1111/j.1540-4781.2011.01146.x) (resumo lido; 661 participantes): relação **linear**, sem limiar mágico, e 98% como alvo mais razoável.
Consequência prática: ler "uma página real" pede milhares de palavras; **textos graduados** (vocabulário controlado) tornam a leitura possível já em T1. Isso resolve o seu objetivo sem esperar anos.

**"Palavra conhecida"** `[PREMISSA]`: lema cujo item receptivo principal está em review com **S ≥ 7 d**. **Cobertura de um texto** = Σ ocorrências de lemas conhecidos ÷ total de tokens (nomes próprios contam como conhecidos).
**Critério de liberação:** ≥ 95% → *leitura guiada* (palavra desconhecida abre glossário PT e vira item novo "vindo da leitura"); ≥ 98% → *leitura livre*. Marcos mostrados ao usuário: "Você conhece 96% deste texto". Esse cálculo usa `reading_text_lexeme` (esquema) e o estado do SRS.

## 4. Nivelamento

Teste **Sim/Não** ([Meara & Buxton 1987](https://doi.org/10.1177/026553228700400202), metadados): o usuário marca as palavras que conhece, misturadas com **pseudopalavras**; os "sim" para pseudopalavras estimam chute e corrigem o resultado `[VERIFICAR conteúdo]`. LexTALE ([Lemhöfer & Broersma](https://doi.org/10.3758/s13428-011-0146-0)) foi desenhado para avançados: não serve para você.
Desenho `[PREMISSA]`: faixas de 500 palavras (`freq_band`), amostra de 8 itens por faixa, começa na faixa 1 e para após 2 faixas seguidas abaixo de 40% (≈ 3 minutos). **Resultado → estado:** as palavras marcadas como conhecidas **não são puladas**: entram como `placement` com S₀ conservador (3 d) e voltam logo para confirmação. Assim o SRS corrige o excesso de confiança sem repetir o que você já sabe.

## 5. Imagens

| Tipo de palavra | Estratégia |
|---|---|
| Concreta (*cat, apple*) | foto/ilustração; vem **depois** da tentativa (Fase 1); some quando S ≥ 21 d |
| Ação/estado (*run, sleep*) | cena ilustrada ou OpenMoji |
| Abstrata (*idea, important*) | frase-contexto + cena; imagem é opcional |
| Função (*the, of, would, although*) | **sem imagem**: lacuna em frase, com 2–3 contextos variados |

| Fonte | Licença | Dá para usar? | Risco / ação |
|---|---|---|---|
| **OpenMoji** | CC BY-SA 4.0; crédito exigido ("All emojis designed by OpenMoji…"); ~4.000 ícones ([FAQ oficial](https://openmoji.org/faq)) | **Sim, base** | Share-Alike: sua coleção derivada fica CC BY-SA |
| **Wikimedia Commons** | por arquivo (CC0, CC BY, CC BY-SA, domínio público) | Sim, com filtro e crédito | Registrar autor + licença por arquivo; **excluir** NC/ND; checar `[VERIFICAR]` direitos de personalidade |
| Pexels / Pixabay | licença própria, não é CC | Não para o acervo | Vedam distribuir "avulso" e montar serviço concorrente ([Pexels](https://help.pexels.com/hc/pt-br/articles/900005880463), [Pixabay](https://pixabay.com/service/license/)); servir um acervo em app tende a esbarrar `[VERIFICAR]` |
| Unsplash API | diretrizes da API | Não | Exige *hotlink* das URLs, crédito com UTM e chamada de *download* ([diretrizes](https://help.unsplash.com/api-guidelines/unsplash-api-guidelines)): inviabiliza o modo offline do PWA |
| IA generativa | depende do modelo/termos; autoria incerta `[VERIFICAR]` | Depois, se couber | custo e revisão; registrar o *prompt* e o modelo em `media` |

**Armazenamento:** arquivos em volume servido pelo proxy (Fase 4), `media` guarda hash, licença e crédito; o repositório só tem o manifesto e o script de carga. Tela de créditos no app lê `media`/`attribution`. **Risco jurídico** (não é aconselhamento): uma imagem com licença errada vira problema do projeto público; por isso `media.status = pending` até alguém aprovar.

## 6. Conexões entre palavras e interferência

[Tinkham 1993](https://doi.org/10.1016/0346-251x(93)90027-e) e [Waring 1997](https://doi.org/10.1016/S0346-251X(97)00013-4) (metadados) apontam que aprender juntas palavras de um mesmo conjunto semântico atrapalha; [Nakata & Suzuki](https://doi.org/10.1017/S0272263118000219) (resumo lido; 133 alunos) acham resultados **inconsistentes** e que o espaçamento pode aliviar o efeito. [Tinkham 1997](https://doi.org/10.1191/026765897672376469) (só título: "semantic and thematic clustering") sustenta a prática de agrupar por **tema**, `[VERIFICAR conteúdo]`.

| Relação | Quando liberar | Observação |
|---|---|---|
| Colocação (*make a decision*) | quando os dois itens têm S ≥ 21 d | vira frase/lacuna, não lista |
| Família (*decide → decision*) | base consolidada primeiro | baixo risco: formas diferentes |
| Sinônimo / antônimo / hiperônimo | **só após S ≥ 21 d dos dois**, em sessões separadas, e como par de contraste | maior risco de interferência |
| Relacionada (tema) | por `unit` temático (comida, rotina…) | agrupar por tema é a regra |

Regra extra `[PREMISSA]`: nunca introduzir na mesma semana dois itens **irmãos semânticos** (mesmo hiperônimo no WordNet: cores, dias, números) — são introduzidos **espalhados**, não em bloco. O grafo guarda `min_stability_days` por aresta. Peso (0–1) = frequência de co-ocorrência no corpus quando existir; senão manual.

## 7. Fontes de conteúdo e licenças

| Fonte | O que dá | Licença | Atenção |
|---|---|---|---|
| **NGSL 1.2** | 2.809 lemas de alta frequência; lista irmã falada, acadêmica e para *graded readers* | Creative Commons Share-Alike; exige citação ([site oficial](https://www.newgeneralservicelist.com/new-general-service-list)); versão **4.0** `[VERIFICAR]` no arquivo baixado | Cite [Browne 2014](https://doi.org/10.7820/vli.v03.2.browne) |
| **Tatoeba** | frases EN↔PT-BR e links | **CC BY 2.0 FR**; parte em **CC0**; áudio com licença por contribuidor ([downloads](https://tatoeba.org/en/downloads)) | Preferir o subconjunto CC0; guardar `source_ref` (id) e autor; compatibilidade 2.0 FR → CC BY-SA `[VERIFICAR]` |
| **Wiktionary** | traduções, classes, sentidos | CC BY-SA `[VERIFICAR]` no dump usado | extrair via dump, não raspar |
| **WordNet / OpenWordNet-PT** | hiperônimos, relações, glosas | licenças próprias `[VERIFICAR]` ([OpenWN-PT](https://repositorio.fgv.br/items/5bff13d2-4ac2-4a02-b46a-e7a5bb9a73b8/full)) | usar só para relações; sem copiar glosas até conferir |
| **PHaVE** (150 *phrasal verbs*; [Garnier & Schmitt 2015](https://doi.org/10.1177/1362168814559798)) e **PHRASE List** ([Martinez & Schmitt 2012](https://doi.org/10.1093/applin/ams010)) | seleção de itens e sentidos mais frequentes | direitos do artigo | **Referência**: escolher os itens por elas e escrever os nossos exemplos |
| **StoryWeaver** | histórias em níveis 1–4 | CC BY 4.0 (conforme o site, por livro `[VERIFICAR]`) | só livros com licença confirmada |
| **Simple English Wikipedia** | textos simples | CC BY-SA `[VERIFICAR]` | vocabulário nem sempre A1-A2 |
| **Oxford 3000/5000** | — | protegidas | só como comparação |

**Licença do projeto:** código MIT (decisão sua); `content/` e o dump do banco em **CC BY-SA 4.0** por herdar Share-Alike de NGSL e OpenMoji, com `NOTICE.md` gerado a partir de `source` e `attribution`. Não é aconselhamento jurídico; antes de publicar o seed, rodar a checagem automática "toda linha tem `source_id` e a licença é permitida". O documento de MIT que você prometeu entra na Fase 4.

**Geração de textos graduados `[PREMISSA]`:** lote offline: (1) lista permitida = lemas do degrau + nomes próprios; (2) gerar por tema; (3) **validador por script** recusa texto com cobertura < 98% da lista; (4) segunda passada automática para naturalidade; (5) amostra revista. Você é A1, então **não consegue** julgar o inglês: por isso a trava automática pesa mais que a revisão humana. Custo no uso: zero. Situação de direitos de texto gerado por IA: `[VERIFICAR]`; publicar como CC BY-SA.

## 8. Modelo de dados (conteúdo)

```mermaid
erDiagram
  subject ||--o{ course : has
  course ||--o{ unit : has
  course ||--o{ lexeme : has
  unit ||--o{ item : groups
  lexeme ||--o{ item : "senses"
  item ||--o{ item_text : "texts per locale"
  item ||--o{ item_media : has
  media ||--o{ item_media : used_in
  item ||--o{ item_relation : from
  item ||--o{ item_relation : to
  source ||--o{ item_text : credits
  source ||--o{ media : credits
  source ||--o{ attribution : credits
  item ||--o| reading_text : is
  reading_text ||--o{ reading_text_lexeme : covers
  lexeme ||--o{ reading_text_lexeme : appears
```

Pontos de projeto: (1) `item.natural_key` normalizada (`word|bank|noun|1`) torna o seed **idempotente** (`ON CONFLICT`). (2) Todo texto traduzível mora em `item_text` com `locale`: qualquer idioma de ajuda novo = novas linhas, sem mudar esquema (separado da UI, que fica em arquivos). (3) `kind` e `exercise` são **extensíveis**: outro assunto só adiciona itens; os exercícios válidos saem de `kind` + configuração, não de tabela. (4) `lexeme` separa lema de sentido e serve à cobertura. (5) Estado do usuário (SRS) **não** está aqui: referencia `item.id` (Fase 3).

## 9. Qualidade e plano do seed (resumo; detalhe na Fase 6)

Pipeline `fonte → JSON/CSV → gerador → SQL idempotente + SQL de conferência`, como o `seed-referencias`. Regras `[PREMISSA]`: deduplicar por `natural_key`; 1 sentido principal por palavra nos 1.000 primeiros (escolha manual dos sentidos 2–3); toda tradução e exemplo com `source_id`; item só vai a `published` com tradução revisada por você e imagem aprovada (concretas). Áudio: **TTS do navegador** (Web Speech API) no MVP, sem arquivos; áudio do Tatoeba só com licença por arquivo.
Escopo realista do MVP de conteúdo: T0 + parte de T1 (**~300–500 palavras**, imagem nas concretas, ~100 frases), em vez das 1.000 iniciais; a Fase 6 confirma.

## 10. Riscos e perguntas em aberto

| Risco | Mitigação |
|---|---|
| Frequência por **sentido** não existe no NGSL (só por lema) | escolher sentidos principais à mão nos 1.000 primeiros; PHaVE para *phrasal verbs* |
| Você não valida o inglês | validador automático + Tatoeba (nativos) + segunda passada |
| Licença errada em imagem/texto | `source`+`attribution`+`status` e checagem no CI |
| Estilo OpenMoji pode parecer infantil | testar com você no Sprint 1; fotos da Commons como alternativa |
| Cobertura "conhecida" por lema ignora sentidos | aceitar no MVP; refinar por sentido depois |

`[DECISÃO SUA]`: (1) `content/` em CC BY-SA 4.0? (2) OpenMoji como imagem base serve, ou prefere fotos? (3) Textos graduados por geração em lote + StoryWeaver, ok? (4) Nivelamento Sim/Não com pseudopalavras, ok?

## 11. Resumo de continuidade (≤ 300 palavras)

Fase 2 concluída (falta seu "ok" e 4 decisões da seção 10). Produto: **wordloop**. Item = 1 sentido/pergunta/resposta; tipos: word, expression, phrasal_verb, sentence, dialogue, paragraph, text. Trilha T0–T4 (≈300 → 1.000 → 2.000 → NGSL 2.809 → listas irmãs); NGSL cobre ~92% de textos gerais (site oficial). Métrica central = **cobertura por texto**: lema conhecido = item receptivo em review com S ≥ 7 d; ≥ 95% libera leitura guiada, ≥ 98% leitura livre (Laufer 2010; Nation 2006; Schmitt 2011: relação linear). Textos graduados: StoryWeaver (CC BY) + geração offline com vocabulário controlado validada por script. Nivelamento: teste Sim/Não com pseudopalavras por faixas de 500; palavras marcadas entram com S₀ = 3 d, não são puladas. Imagens: OpenMoji (CC BY-SA 4.0) como base, Wikimedia Commons filtrado; Pexels/Pixabay/Unsplash descartados; arquivos **fora do Git** com manifesto `media` (hash, licença, crédito, status). Função/abstratas: sem imagem (lacuna). Conexões em grafo tipado, liberadas só com S ≥ 21 d nos dois itens; nunca irmãos semânticos na mesma semana; unidades por tema. Licenças: código MIT; `content/` CC BY-SA 4.0 com NOTICE; Tatoeba (CC BY 2.0 FR/CC0) e Wiktionary `[VERIFICAR]`. Modelo: subject→course→unit→item, `lexeme`, `item_text` por locale, `item_relation`, `media`, `source`, `attribution`, `reading_text(_lexeme)`; DDL em `docs/fase-2/content-schema.sql`. Áudio: TTS do navegador. MVP de conteúdo: ~300–500 palavras. Próxima: Fase 3 (funcionalidades, arquitetura da API e do front, modelo de dados completo, contrato de API, identidade e push).
