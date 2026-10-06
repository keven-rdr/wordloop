# Fase 1 — Repetição espaçada (SRS)

> Data: 2026-10-06. Rótulos: `[PREMISSA]` assumi; `[VERIFICAR]` não confirmei em fonte primária; `[DECISÃO SUA]` só você decide.
> **Como li as fontes:** conferi título/autor/ano/periódico de cada DOI no Crossref. Li o resumo (abstract) só dos itens marcados "resumo lido". Nos demais, o conteúdo vem da sua descrição ou da minha memória: `[VERIFICAR]`.
> Código de apoio, reproduzível: [srs_reference.py](srs_reference.py) (FSRS-6 + regra simples) e [worked_example.py](worked_example.py).

## 1. Tabela de decisões

| Decisão | Recomendação | Alternativas descartadas | Por quê |
|---|---|---|---|
| "% de conhecimento" | **Dois números**: *domínio* (derivado de S; muda só quando você responde) é o que a UI mostra por item; *R* (retrievability, cai com o tempo) é calculado na hora e usado só para ordenar fila e escolher a pergunta da notificação | só R (UI mostraria 100% → 62% sem você fazer nada); só escore próprio | R é a probabilidade estimada de lembrar agora e tem sentido estatístico; mostrá-la por item assusta e desmotiva. Domínio é estável e motivador |
| Curva de esquecimento | Uma função única `R(t,S)` (curva de potência do FSRS-6) compartilhada por **todas** as estratégias | cada estratégia com sua curva | Estratégias só diferem em *como atualizar S*; assim o % e os gráficos são comparáveis |
| Estratégias | `simple_v1` (transparente, ajustável) no MVP; `fsrs6` como segunda implementação; mesma interface | só FSRS; só regra simples | `go-fsrs` já existe, então FSRS custa pouco; mas a regra simples é o seu baseline didático e de teste |
| Nota | Mapeamento do seu rascunho **validado** com duas regras extras (seção 4.4) | usar só o que o usuário toca | A nota subjetiva é enviesada (Koriat & Bjork 2005) |
| Autoavaliação de acerto | **Não existe**: o app corrige a resposta (digitar/escolher). A dificuldade só *modula* | o usuário julgar se acertou (como no Anki) | Elimina a "mentira" que o seu arquivo aponta ("clicar Fácil para pular") |
| Unidade agendada | `(usuário, sentido, direção)`; **receptivo (EN→PT/ouvir) primeiro**, produtivo (PT→EN) libera depois `[PREMISSA]` | um estado por palavra | Seu gargalo é compreensão; recepção e produção têm forças de memória diferentes |
| Dia | Vira às **04:00 no fuso do usuário** `[PREMISSA]` (como no Anki, ver 4.7) | meia-noite | Estudo noturno não "quebra" a sequência |

## 2. Evidência

| Achado | Fonte | Força | Impacto no produto |
|---|---|---|---|
| Espaçar melhora a retenção; o intervalo ótimo **cresce** com o prazo de retenção, mas **cai em proporção** (≈20–40% de 1 semana → ≈5–10% de 1 ano) | [Cepeda et al. 2008](https://doi.org/10.1111/j.1467-9280.2008.02209.x) (resumo lido; >1.350 pessoas); [Cepeda et al. 2006](https://doi.org/10.1037/0033-2909.132.3.354) (meta-análise; só metadados) | Forte | Intervalos devem crescer com S; sem tabela fixa |
| Vocabulário de L2: 13 sessões a cada 56 dias ≈ 26 a cada 14 dias | [Bahrick et al. 1993](https://doi.org/10.1111/j.1467-9280.1993.tb00571.x) (resumo lido: **4 participantes**, 9 anos) | Sugestiva, amostra mínima | Não usar como número mágico |
| Testar > reestudar quando o teste final é adiado | [Roediger & Karpicke 2006](https://doi.org/10.1111/j.1467-9280.2006.01693.x) (resumo lido); [Karpicke & Roediger 2008](https://doi.org/10.1126/science.1152408) (metadados) | Forte | Todo estudo é **recuperação ativa**; nada de "só olhar" |
| Expansivo × igual: o expansivo ganha no curto prazo; o igualmente espaçado ganha em 2 dias; o que importa é **atrasar a 1ª recuperação** | [Karpicke & Roediger 2007](https://doi.org/10.1037/0278-7393.33.4.704) (resumo lido). Em L2, [Nakata 2015](https://doi.org/10.1017/S0272263114000825): vantagem do expansivo de ~4,6% só com correção branda, sem diferença com correção estrita (resultado de busca, `[VERIFICAR]` no artigo) | **Controversa** | Não otimizar o formato da expansão; **1ª revisão ≥ próximo dia**; passos curtos só para corrigir erro |
| A curva de esquecimento de Ebbinghaus se replica; há um "salto" por volta de 24 h | [Murre & Dros 2015](https://doi.org/10.1371/journal.pone.0120644) (resumo lido; 1 sujeito) | Moderada | Justifica a 1ª revisão em ~1 dia |
| Julgamentos de aprendizado têm **ilusão de competência** | [Koriat & Bjork 2005](https://doi.org/10.1037/0278-7393.31.2.187) (resumo lido) | Forte | Dificuldade informada ≠ verdade: combinar com acerto e tempo |
| Heurísticas com parâmetros fixos são inferiores a algoritmos treináveis com dados | [Tabibian et al. 2019](https://doi.org/10.1073/pnas.1815156116) (resumo lido; Duolingo); [Settles & Meeder 2016](https://doi.org/10.18653/v1/P16-1174) | Moderada (dados de plataforma) | Registrar log completo para treinar depois |
| Imagem ajuda quando há recuperação ativa ou aviso; sozinha, gera excesso de confiança | [Carpenter & Olson 2012](https://doi.org/10.1037/a0024828); [Chun & Plass 1996](https://doi.org/10.1111/j.1540-4781.1996.tb01159.x) (só metadados; conteúdo = sua descrição) `[VERIFICAR]` | Moderada | Política de exercícios (seção 9); aprofundo na Fase 2 |
| Agrupar palavras muito parecidas atrapalha; resultados são inconsistentes e o espaçamento pode aliviar | [Tinkham 1993](https://doi.org/10.1016/0346-251x(93)90027-e) (metadados); [Nakata & Suzuki 2018/19](https://doi.org/10.1017/s0272263118000219) (resumo lido: 133 alunos) | Controversa | Perguntas "relacionadas" = mesmo item, não vizinhos semânticos |
| "Dificuldades desejáveis" (Bjork & Bjork) | capítulo sem DOI | `[VERIFICAR]` | Está embutido na fórmula do FSRS: lembrar com R baixo aumenta mais S |

**Estabelecido:** espaçamento, efeito do teste, recuperação ativa. **Controverso:** forma da expansão, benefício exato da imagem, interferência semântica. **Em português** (ver seção 12): a literatura é de relatos e revisões, com evidência fraca de causalidade; serve de motivação, não de prova.

## 3. Algoritmos para a *sua* regra (3 níveis + %)

| | Leitner | SM-2 (Anki clássico) | HLR (Duolingo) | FSRS-6 |
|---|---|---|---|---|
| Modelo | caixas | fator de facilidade (EF) por cartão | meia-vida por regressão | DSR: S, D e R explícitos |
| Dá "% de conhecimento"? | não | não | sim | **sim** (R) |
| Usa 3 níveis? | não | 4 notas (0–5 no original) | não (acerto/erro + contagens) | 4 notas |
| Precisa de dados? | não | não | muitos (treino no servidor) | funciona com padrões; otimiza com histórico |
| Evidência | histórica | heurística de 1987 | [Settles & Meeder 2016](https://doi.org/10.18653/v1/P16-1174) | [Ye et al. KDD 2022](https://doi.org/10.1145/3534678.3539081); [Su et al. TKDE 2023](https://doi.org/10.1109/TKDE.2023.3251721) |
| [srs-benchmark](https://github.com/open-spaced-repetition/srs-benchmark) (~1,7 bi de revisões, 20 mil usuários; benchmark da comunidade, não revisado por pares) | — | não listado no resumo que li `[VERIFICAR]` | log loss 0,4694 | log loss 0,3460 (menor é melhor) |

Leitura: o FSRS é melhor em dados reais, mas a vantagem foi medida em usuários do Anki, não nos seus. Ele é a melhor aposta para o % de conhecimento; a regra simples fica como baseline explicável. Biblioteca Go: [go-fsrs](https://github.com/open-spaced-repetition/go-fsrs) (MIT; módulo `/v4`; implementa FSRS-6 segundo o README `[VERIFICAR]`).

## 4. Especificação do agendador

### 4.1 Grandezas
- **S** (estabilidade, dias): tempo em que R cai de 100% para 90%.
- **R(t,S) = (1 + F·t/S)^(−d)**, com `d = 0,1542` (w20) e `F = 0,9^(−1/d) − 1`, logo R(S,S) = 90% (fórmula do [wiki do FSRS](https://github.com/open-spaced-repetition/awesome-fsrs/wiki/The-Algorithm)).
- **Intervalo** para retenção desejada `r`: `I = S/F · (r^(−1/d) − 1)`; com r = 0,9, I = S. Mínimo 1 dia, máximo 365 dias no MVP `[PREMISSA]`.
- **Domínio** = `clamp(ln(1+S) / ln(1+180), 0, 1)` `[PREMISSA: 180 d, calibrar]`. S = 2 → 21%; 9 → 44%; 30 → 66%; 90 → 87%.
- **Dominado**: estado Review com S ≥ 90 d e sem lapso nas 2 últimas revisões `[PREMISSA]`. Continua voltando, com intervalos longos (o seu "100% ≠ nunca mais").

### 4.2 Estados
`new → learning → review ⇄ relearning`; "dominado" é derivado (S ≥ 90) e `leech` é uma marca, não um estado.
- **Learning** (1ª exposição; o app *ensina* antes de testar, ao contrário do Anki): apresentação sem nota, depois teste de reconhecimento na mesma sessão; passos de **3–5 cartões depois** e depois ~10 min. Resposta correta no passo final gradua para review com S₀ da nota.
- **Relearning**: após erro em review, passos curtos na mesma sessão; o erro já reduziu S (fórmula abaixo).

### 4.3 Regras de atualização
**FSRS-6** (parâmetros padrão do wiki; D ∈ [1,10]):

| Caso | Fórmula |
|---|---|
| S inicial | `S₀(G) = w[G−1]` → Again 0,21 · Hard 1,29 · Good 2,31 · Easy 8,30 |
| D inicial | `D₀(G) = w4 − e^{w5(G−1)} + 1` |
| Acerto | `S' = S·(1 + e^{w8}·(11−D)·S^{−w9}·(e^{w10(1−R)} − 1)·h·e)`, `h = w15` (Hard), `e = w16` (Easy) |
| Erro | `S' = min(S, w11·D^{−w12}·((S+1)^{w13} − 1)·e^{w14(1−R)})` |
| Mesmo dia | `S' = S·e^{w17(G−3+w18)}·S^{−w19}` |
| D | `ΔD = −w6(G−3)`; `D' = D + ΔD·(10−D)/9`; regressão à média com w7 |

Reproduzi a tabela do seu arquivo (respondendo sempre *Good*): S = 2,3 → 10,9 → 46,2 → 162,7 → 496,5; o arquivo diz 11,0 / 46,3 / 163 / 498 (diferença < 0,5%, arredondamento).

**`simple_v1`** (cada número cabe numa tabela e você pode ajustá-lo):

| Situação | Regra |
|---|---|
| 1º acerto | S = 1 (Difícil) · 2 (Médio) · 4 (Fácil) dias |
| Acerto | `S' = (S + 0,5·atraso) · m`, `m` = 1,2 (Difícil) · 2,0 (Médio) · 3,0 (Fácil) |
| Erro | `S' = max(1, 0,25·S)` e `lapsos += 1` |
| Limites | S ≤ 365; fuzz como em 4.7 |

Valores de m e do fator 0,25 são `[PREMISSA]`, a calibrar por simulação (seção 7).

### 4.4 Combinar acerto, dificuldade e tempo
| Resultado | Dificuldade informada | Nota |
|---|---|---|
| errou | qualquer | **Again** |
| acertou | Difícil | **Hard** |
| acertou | Médio | **Good** |
| acertou | Fácil | **Easy** |

Seu mapeamento está correto. Duas regras minhas `[PREMISSA]`, a validar com dados:
1. **Easy só vale em exercício de produção** (digitar, ouvir→digitar). Em reconhecimento (múltipla escolha) "Fácil" conta como Good, pois escolher é mais fácil que lembrar e infla S. Isso responde à ilusão de competência.
2. **Dica, explicação ou imagem vistas antes de responder** limitam a nota a Hard.

**Tempo de resposta não altera a nota.** Ele só pré-seleciona a sugestão: ≤ 0,6× a sua mediana para aquele tipo de exercício → Fácil; ≥ 1,5× → Difícil; senão Médio. Você pode trocar. Guardo `sugerido` e `escolhido` para medir o viés de ancoragem depois.

**Quando erra:** a pergunta de dificuldade não faz sentido ("foi fácil lembrar?"). Redação para o erro: **"Como foi?"** → *Não sabia* · *Quase lembrei* · *Foi distração*. A nota continua Again, mas: (a) *Foi distração* ou erro de digitação com distância de edição ≤ 1 (palavras ≥ 5 letras) vale **Hard sem lapso** `[DECISÃO SUA]`; (b) a resposta define o passo de reaprendizado (*Não sabia* volta em 3 cartões e mostra a imagem de contexto). Para acerto, redação: **"Foi fácil, médio ou difícil lembrar?"**

### 4.5 Lapsos, sanguessugas, atrasos, limites
- **Lapso** só em review/relearning. **Sanguessuga**: ≥ 6 lapsos `[PREMISSA]` (Anki: 8 `[VERIFICAR]`). Ação: **nunca suspender em silêncio**; oferecer trocar imagem/frase-contexto ou reescrever.
- **Atraso**: FSRS já embute o crédito via R; no `simple_v1`, 50% do atraso.
- **Novos por dia**: derivado da meta de minutos do dia: `novos ≈ minutos / 5` (60 min → 12) `[PREMISSA]`, com teto 20 e **zero novos se a fila de revisão passa de 2× a meta** (proteção contra avalanche). Revisões: sem teto, ordenadas por **menor R**.
- **Meta dinâmica:** os limites são recalculados quando o usuário define a meta no início do ciclo.

### 4.6 Fila e notificação-desafio
Fila = `learning` vencido → `relearning` → `review` por menor R → novos. A pergunta da notificação sai de um item **vencido ou em aprendizado**, no formato mais curto (lacuna ou ouvir→escolher), e é registrada com `source = notification`.

### 4.7 Dia, fuzz, irmãos
- **Dia lógico** = `floor((agora_local − 4 h) / 24 h)` no fuso do usuário; o fuso e o corte ficam no perfil e **dentro de cada linha do log** (o fuso muda).
- **Fuzz** (evitar pilhas no mesmo dia): nenhum abaixo de 2,5 d; ±15% de 2,5 a 7 d; ±10% de 7 a 20 d; ±5% acima, como no Anki `[VERIFICAR no código rslib]`. Sorteio determinístico por `hash(card_id, n_revisão)` para ser testável.
- **Irmãos** (receptivo × produtivo do mesmo sentido): o produtivo só nasce quando o receptivo tem S ≥ 7 d; nunca na mesma sessão; entre irmãos, ≥ 1 dia de distância.

## 5. Strategy

```go
type Scheduler interface {
    Name() string                                      // "simple_v1", "fsrs6"
    Next(c CardState, r Review, now time.Time) (CardState, error) // função pura
    Retrievability(c CardState, now time.Time) float64
}
```
Registro por nome (`user_settings.scheduler`). O estado da carta guarda S/D/estado/vencimento/lapsos; parâmetros específicos ficam em `memory jsonb`. **Trocar de estratégia = reproduzir o log imutável** do usuário na nova implementação (`replay(reviews) → CardState`), o que torna a comparação possível. O padrão `supports`/`handle` do Contratos não é necessário aqui: a escolha é por configuração, não por evento.

## 6. Exemplo passo a passo (item *sort*; 6 respostas)

Respostas: Médio ✔ · Fácil ✔ · Difícil ✔ com 3 dias de atraso · erro · Médio ✔ · Médio ✔. Cada algoritmo segue a **sua** data de vencimento ([worked_example.py](worked_example.py)):

| # | FSRS-6 (retenção 90%) | `simple_v1` |
|---|---|---|
| 1 | dia 0, Good: S = 2,3; próxima em 2 d | dia 0: S = 2; 2 d; domínio 21% |
| 2 | dia 2, Easy: R = 0,91 → S = **18,5**; próxima em **19 d** | dia 2, Fácil: S = 6; 6 d; 37% |
| 3 | dia 24 (3 d atrasado), Hard: R = 0,89 → S = 60,2; 60 d | dia 11 (3 d atrasado), Difícil: S = 9,0; 9 d; 44% |
| 4 | dia 84, **erro**: S = 3,1; 3 d | dia 20, **erro**: S = 2,2; 2 d; 23% |
| 5 | dia 87, Good: S = 7,2; 7 d | dia 22, Médio: S = 4,5; 4 d; 33% |
| 6 | dia 94, Good: S = 15,3; 15 d | dia 26, Médio: S = 9,0; 9 d; 44% |

**Lição que muda o produto:** com parâmetros padrão e **um** "Fácil" no 2º contato, o FSRS salta para 19 dias e depois 60. Num iniciante com nota subjetiva viesada isso é arriscado. Mitigações: regra 1 da seção 4.4, `r = 0,90` no começo (subir para 0,92–0,95 só se a simulação mostrar lapsos altos), e **teto de crescimento por revisão** (ex.: I' ≤ 3× o intervalo anterior nas primeiras 5 revisões do item) `[PREMISSA]`. O FSRS pode ser otimizado com o seu histórico; antes disso, confie menos nos padrões.

## 7. Plano de testes

| Camada | O que testar |
|---|---|
| Golden vectors | Sequências fixas conferidas com [srs_reference.py](srs_reference.py) e, quando possível, com `go-fsrs`/`py-fsrs` `[VERIFICAR]`; tolerância 1e-6 |
| Propriedades (ex.: `pgregory.net/rapid` `[VERIFICAR]`) | S > 0; 0 < R ≤ 1; R decresce com t; R(S,S) = 0,9; intervalo(Easy) ≥ Good ≥ Hard ≥ Again; vencimento > agora; lapso só incrementa em erro; D ∈ [1,10]; sem NaN/Inf para entradas extremas; **determinismo** (mesma entrada, mesma saída) |
| Replay | `replay(log) == estado incremental`; reenvio do mesmo UUID não altera nada (idempotência) |
| Tempo | virada de dia por fuso, mudança de horário de verão, relógio do cliente adiantado/atrasado |
| Simulação | Alunos virtuais com esquecimento de verdade (verdade ≠ modelo, para não avaliar o algoritmo por ele mesmo). Medir: retenção real nas revisões vs. alvo, revisões/dia, itens com R ≥ 0,9 no dia 30/90, **calibração** (R previsto × acerto observado em faixas, como o RMSE bins do benchmark) |

## 8. Log imutável de revisões (`review_log`, só INSERT)

Campos: `id` (UUID **gerado no cliente**) · `user_id` · `card_id` · `sense_id` · `exercise_type` · `source` (session / notification / placement) · `shown_at`/`answered_at` (cliente) + `received_at` (servidor) · `tz`/`day_cutoff` · `response_ms` · `is_correct` · `answer_raw` · `typo_flag` · `hint_used` · `explanation_viewed` · `image_before_answer` · `difficulty_suggested` · `difficulty_chosen` · `grade` (1–4) · **estado antes**: `state`, `S`, `D`, `R`, `elapsed_days` · **estado depois**: `state`, `S`, `D`, `scheduled_days`, `due_at` · `scheduler` + `version` + `params_id` · `device_id` · `app_version` · `client_seq`.

Imutável = a aplicação só insere; correções entram como evento `void` que referencia o original (por exemplo, perguntas com gabarito errado). Exclusão de conta (LGPD) anonimiza ou apaga, e isso é uma exceção documentada. Esses campos bastam para o otimizador do FSRS (ordem, nota, intervalo decorrido) e para HLR.

## 9. Política de exercícios

| Exercício | Quando | Papel da imagem |
|---|---|---|
| Palavra EN → escolher PT (reconhecimento) | **Novos** e 1ª revisão | nenhum antes da resposta; aparece depois |
| Imagem → digitar/escolher a palavra | revisão de concretas; recuperação ativa com a imagem como *dica*, não resposta | cue |
| Palavra → escolher imagem | só reforço leve; pode virar muleta | resposta (evitar como principal) |
| Ouvir (TTS) → escolher / digitar | desde cedo (seu gargalo é compreensão auditiva) | opcional |
| PT → digitar EN (produção) | quando S ≥ 7 d | **oculta** |
| Lacuna (*cloze*) | frases; palavras abstratas e funcionais (*the, of, would*) | contexto, não figura |
| Montar frase | depois que as palavras estão em review | nenhum |
| Múltipla escolha × digitar | escolha para introdução; **digitar para S ≥ 7 d** | — |

**Contra a muleta:** a imagem aparece **depois** da tentativa por padrão (coerente com Carpenter & Olson); *fading*: a partir de S ≥ 21 d some do exercício. A explicação e "por que errei" (seu pedido) mostram a imagem como contexto.

## 10. Riscos e perguntas em aberto

| Risco | Mitigação |
|---|---|
| Padrões do FSRS otimistas para iniciante | teto de crescimento; simulação; otimizar depois de ~1.000 revisões `[VERIFICAR]` |
| Nota subjetiva enviesada | regras 4.4; guardar sugerido × escolhido |
| 12 itens novos/dia pode ser alto para A1 | começar em 8 e ajustar pela taxa de lapsos `[DECISÃO SUA]` |
| Receptivo-primeiro atrasa a produção | revisitar com dados |

`[DECISÃO SUA]`: (1) *Foi distração* vale Hard sem lapso? (2) "Dominado" = S ≥ 90 d serve? (3) A regra Easy-só-em-produção é aceitável?

## 11. Sobre o seu arquivo (BASE TEORICA…)

- **Código e números do FSRS/SM-2:** reproduzi a tabela de S e as fórmulas do FSRS-6 (acima). Fuzz, sanguessuga e deltas do EF do Anki conferem com o que lembro do `rslib`, mas **não li o código**: `[VERIFICAR]`.
- **"% de conhecimento = R, derivado e não armazenado"**: correto, e é a base da seção 4.1.
- **Fontes (verificadas hoje):** Cruccioli et al. 2024 existe ([RESU](https://revistas.unievangelica.edu.br/index.php/educacaoemsaude/article/download/7479/5461/33486); 20 estudos, 2.029 estudantes; desempenho melhor em 12 e indiferente em 8; **não conseguiu dizer se o método é superior**). O arquivo afirma que "comprovadamente" reduz ansiedade e melhora o sono, mas isso vem de relatos dos alunos: **afirmação forte demais**. Fernandes 2022 ([DOI](https://doi.org/10.33448/rsd-v11i4.27347)) e Gilbert 2023 ([DOI](https://doi.org/10.1007/s40670-023-01826-8)) existem. Ventura 2021 existe, mas o autor no Crossref é *Ventura Gonçalves* ([DOI](https://doi.org/10.35168/2176-896x.utp.tuiuti.2021.vol7.n63.pp131-149)), não "Ventura e Santos". **Costa (UFRN, 2024): não encontrei**; não cite até achar o link.
- "O Anki não ensina": nosso app **ensina e depois testa** (seção 4.2). "Cartão atômico": vira regra de conteúdo (um sentido, uma pergunta) na Fase 2.

## 12. Referências em português e extras

[Haraki 2023](https://repositorio.ufsc.br/handle/123456789/248930) (TCC UFSC, revisão sistemática); [Silva & Rodrigues 2022](https://repositorio.ifap.edu.br/items/c4f44eaf-9ade-4d2f-ae4a-a60429d1b52e/full) (TCC IFAP); [Rios 2021](https://ubibliorum.ubi.pt/bitstreams/0fdec4ed-c223-4118-922f-3d878965a15c/download) (dissertação UBI, Portugal; espaçamento e efeito de teste); [Pereira 2024](https://dspace.bc.uepb.edu.br/jspui/bitstream/123456789/32375/1/TCC-%20walison%20Pereira%20finalizado.pdf) (TCC UEPB sobre o Anki; apareceu na busca, não li); [UFSCar: repetição espaçada e memória de longo prazo](https://repositorio.ufscar.br/handle/ufscar/17354) (apareceu na busca, não li). Nenhuma delas é evidência causal para o seu caso. Não tive acesso a SciELO/BDTD/CAPES por busca direta: `[VERIFICAR]` em outra rodada se você quiser mais.
Outras: [Murre & Dros](https://doi.org/10.1371/journal.pone.0120644); [Dunlosky et al. 2013](https://doi.org/10.1177/1529100612453266) e [Pimsleur 1967](https://doi.org/10.1111/j.1540-4781.1967.tb06700.x) (só metadados); Leitner 1972 e Woźniak (SM-2, 1990) sem DOI `[VERIFICAR]`; [Paivio 1991](https://doi.org/10.1037/h0084295) (metadados).

## 13. Resumo de continuidade (≤ 300 palavras)

Fase 1 (SRS) concluída; falta o seu "ok". Decisões: (1) **dois números**: *domínio* = ln(1+S)/ln(181), mostrado por item e alterado só ao responder; *R* = (1+F·t/S)^−0,1542, calculado na hora e usado para fila e notificação. (2) Interface `Scheduler` (`Next` pura + `Retrievability`) com `simple_v1` (Hard×1,2/Médio×2/Fácil×3; erro ×0,25; 1º acerto 1/2/4 d) no MVP e `fsrs6` (go-fsrs v4, MIT) depois; trocar = replay do log imutável. (3) Nota: errou→Again; acertou+Difícil/Médio/Fácil→Hard/Good/Easy; Easy só em produção; dica/imagem antes limitam a Hard; tempo só sugere a dificuldade (≤0,6×mediana→Fácil, ≥1,5×→Difícil); no erro a pergunta vira "Não sabia / Quase / Foi distração". (4) Estados new/learning/review/relearning; dominado = S≥90 d; sanguessuga ≥6 lapsos (oferece trocar imagem/frase); dia vira às 04:00 no fuso do usuário; fuzz como no Anki; novos/dia ≈ minutos/5, zero novos se fila > 2× meta. (5) Unidade = (usuário, sentido, direção); receptivo primeiro, produtivo com S≥7 d. (6) Exercícios: reconhecimento para novos, digitar com S≥7 d, imagem depois da tentativa e some com S≥21 d, cloze para palavras funcionais, TTS cedo. (7) `review_log` só-INSERT com UUID do cliente, estado antes/depois, sugerido×escolhido, scheduler+versão. Descoberta: FSRS com padrões salta para 19 d após um "Fácil" no 2º contato; mitigar com teto de crescimento. Evidência: espaçamento e teste fortes; expansão e imagem controversas. Pendente: 3 decisões da seção 10 e os fatos do seu arquivo (Costa UFRN não encontrado; Cruccioli superestimado). Próxima: Fase 2 (pedagogia, conteúdo e modelo de dados de conteúdo).
