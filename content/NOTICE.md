# Avisos e créditos do conteúdo

O **conteúdo** de `content/` e das migrações `api/migrations/*content*` (palavras, glosas, frases, exemplos) é distribuído sob **CC BY-SA 4.0**, e o **código** sob MIT (ver [ADR 0024](../docs/adr/0024-licencas-codigo-mit-e-conteudo-cc-by-sa-4-0.md)). O conteúdo reúne material de terceiros, com as licenças abaixo. **Isto não é aconselhamento jurídico**; a compatibilidade entre licenças aguarda revisão (ver [ADR 0031](../docs/adr/0031-fontes-do-conteudo-v1.md)).

| Fonte | Uso | Licença |
|---|---|---|
| Textos originais do projeto wordloop | glosas em pt-BR, seed v0 | CC BY-SA 4.0. As glosas e traduções são **assistidas por IA e não passaram por revisão humana** |
| [New General Service List (NGSL) 1.2](https://www.newgeneralservicelist.com/), de Browne, Culligan e Phillips | lista e ordem de frequência dos lemas do lote v1 | CC BY-SA 4.0 |
| [Tatoeba](https://tatoeba.org/) | frases em inglês, suas traduções em português e os exemplos de cada palavra | CC BY 2.0 FR |
| [OpenMoji](https://openmoji.org/) | imagens (`web/public/media/openmoji`) | CC BY-SA 4.0 |

## Atribuição do Tatoeba
Cada frase e cada exemplo do lote v1 guarda o número das duas frases de origem (`eng_id` e `por_id` em `content/seed/v1/*.json`; no banco, `item_text.source_ref` no formato `eng#<id>/por#<id>`) e o nome dos autores (`eng_user` e `por_user`; no banco, a tabela `attribution`). A página de uma frase é `https://tatoeba.org/sentences/show/<id>`. A tela de créditos do app lista as fontes e a licença de cada uma.

## Como foi gerado
`content/tools/fetch_sources.py` baixa as fontes (url, data e sha256 em `content/seed/v1/sources.json`), `build_v1.py` escolhe palavras e frases a partir delas e das glosas em `content/seed/v1/glosses/`, `validate.py` confere, e `gen_v1.py` gera as migrações. Os arquivos baixados não são versionados.
