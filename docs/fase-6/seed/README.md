# Seed de conteúdo — amostra e plano

Pipeline reproduzível: **fonte JSON → `gen_sql.py` → SQL idempotente + SQL de conferência**, o mesmo padrão do `seed-referencias` do Contratos (escrito do zero, sem código de lá).

## Como rodar (amostra)

```bash
cd docs/fase-6/seed
python gen_sql.py                 # gera out/sample-seed.sql e out/sample-verify.sql
python validate_sqlite.py         # valida num SQLite em memória (integridade, CHECKs, idempotência, conferência)
```
No **PostgreSQL** (ainda **não executado**; é o primeiro item do Sprint 0 do banco):
```bash
psql "$DATABASE_URL" -f ../../fase-2/content-schema.sql   # esquema de conteúdo
psql "$DATABASE_URL" -f out/sample-seed.sql               # carga (pode repetir)
psql "$DATABASE_URL" -f out/sample-verify.sql             # contagens = 'ok'; violações = 0 linhas
```

## O que a amostra contém

| Tipo | Quantidade |
|---|---|
| Palavras (`word`), com 2 sentidos de *sort* (verbo e substantivo) | 31 |
| Expressões / colocações | 7 |
| *Phrasal verbs* | 3 |
| Frases (com lacuna `{{...}}`) | 10 |
| Relações (antônimo, colocação, relacionada) | 8 |
| Imagens (OpenMoji, `pending`) | 10 |
| Linhas de texto (`item_text`) | 194 |

Resultado da validação de hoje (SQLite): contagens batem, idempotente (2 execuções = mesmas contagens), saída determinística (mesmo JSON = mesmo SQL), e um **teste negativo** (dados sabotados) é detectado pela conferência.

## Pontos de atenção

1. **Idempotência:** `INSERT … ON CONFLICT` na chave natural; IDs são UUID v5 derivados da chave natural (mesma entrada, mesmo id).
2. **Não sobrescreve decisão humana:** `status` e `content_version` **não** entram no `DO UPDATE`; reexecutar o seed não "despublica" nem reverte revisão.
3. **Tudo `draft`:** nenhuma tradução da amostra foi revisada. Os textos são **originais do projeto**, escritos para validar o modelo (fonte `wordloop-original`, CC BY-SA 4.0). Direitos de texto gerado com IA: `[VERIFICAR]`.
4. **Imagens:** só o manifesto. `sha256` é o marcador `pending:calcular-na-ingestao`, `status = pending`; a ingestão baixa o arquivo, calcula o hash e só então vira `approved`. Códigos hexadecimais do OpenMoji: `[VERIFICAR]` contra o acervo.
5. **Frequência e CEFR:** `freq_rank`, `freq_band` e `coverage_share` ficam **nulos** na amostra (não confirmei o ranking contra o arquivo do NGSL); o CEFR é **julgamento meu**, `[PREMISSA]`.
6. **O que não foi provado:** a sintaxe PostgreSQL. A validação foi em SQLite (mesmo SQL, DDL adaptado).

## Plano do seed completo (MVP: ~500 palavras)

| Etapa | O que faz | Saída |
|---|---|---|
| 1. Aquisição | baixar NGSL 1.2, Tatoeba (subconjunto CC0 e CC BY 2.0 FR), dump do Wiktionary; **uma linha em `source` por fonte** com versão, licença e `sha256` do arquivo | `source/raw/` (fora do Git; só o manifesto entra) |
| 2. Seleção | lemas por faixa de frequência (NGSL), com `coverage_share`; **sentido principal escolhido à mão** nas primeiras 500 palavras | `source/words.json` |
| 3. Frases | do Tatoeba: tem tradução PT-BR, ≤ 8 palavras, **todas as palavras já ensinadas + no máximo 1 nova**, prioridade para as bem avaliadas | `source/sentences.json` |
| 4. Traduções | do Tatoeba/Wiktionary; o que faltar é rascunho assistido, marcado `draft` | `review/to-review.csv` |
| 5. Revisão humana | **você** revisa PT-BR em planilha; o inglês passa por validação automática (vocabulário permitido) e segunda leitura | `review/reviewed.csv` → `status = reviewed` |
| 6. Imagens | OpenMoji por mapeamento palavra→hexcode (manual nas concretas); Wikimedia Commons só com licença permitida | `source/media.json` |
| 7. Geração | `gen_sql.py` → seed + conferência | `out/*.sql` |
| 8. Carga | comando `api seed` (idempotente) no deploy de tst; **prd só depois da revisão** | — |

**Metas do MVP** `[PREMISSA]`: ~500 palavras (T0 + parte de T1), ~100 frases, ~40 expressões e *phrasal verbs*, imagem em ~150 concretas, ~100 colocações e ~30 antônimos como relações. **Regras de qualidade:** sem duplicata por `natural_key`; 1 sentido principal por palavra nos 500 primeiros; todo texto com `source_id`; `published` só com tradução revisada e (nas concretas) imagem `approved`; checagem automática de licença no CI.
