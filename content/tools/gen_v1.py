#!/usr/bin/env python3
"""Gera as migracoes do lote v1 a partir de content/seed/v1/*.json (depois de validate.py).

Uso:  python content/tools/gen_v1.py        (a partir da raiz do repositorio)
Saida: api/migrations/00020_content_v1_base.sql, 00021_content_v1_words.sql, 00022_content_v1_sentences.sql
       e 00023_content_publish.sql (publica o seed v0).

Diferenca para o gen_sql.py do v0: INSERT com VARIAS linhas por instrucao (500). Uma linha por instrucao custa ~0,15 s
de ida e volta no Neon (as 849 do v0 ja estouraram um limite de 120 s); em lote sao dezenas de instrucoes.
Idempotente (ON CONFLICT DO NOTHING) e deterministico (UUID v5 pela chave natural).
"""
import json
import sys
import uuid

from common import MIGRATIONS, ROOT, V1, norm, q
from validate import load, validate

BATCH = 500
COURSE_SLUG = "en-for-pt-br"
SEED_V0 = json.loads((ROOT / "content" / "seed" / "seed-v0.json").read_text(encoding="utf-8"))
NS_V0 = uuid.uuid5(uuid.NAMESPACE_URL, SEED_V0["namespace"])
V0_LEXEMES = {(norm(w["lemma"]), w["pos"]) for w in SEED_V0["words"]} | {(norm(e["lemma"]), e["pos"]) for e in SEED_V0["lexemes_extra"]}
NS = uuid.uuid5(uuid.NAMESPACE_URL, "wordloop-seed-v1")
COURSE_ID = str(uuid.uuid5(NS_V0, f"course|{COURSE_SLUG}"))
FUNCTION_POS = {"det", "prep", "conj", "pron"}
HEADER = "-- GERADO por content/tools/gen_v1.py a partir de content/seed/v1/*.json. NÃO EDITE À MÃO.\n"
DOWN = "-- +goose Down\n-- Sem desfazer: o conteúdo é só acrescentado (expand); retirar conteúdo é decisão de contract futura.\nSELECT 1;\n"


def uid(*parts: str) -> str:
    return str(uuid.uuid5(NS, "|".join(parts)))


def lexeme_id(lemma: str, pos: str) -> str:
    """Lexema que o v0 ja criou (ex.: `look`, usado pelos phrasal verbs) mantem o id do v0; os demais ganham id do v1."""
    if (lemma, pos) in V0_LEXEMES:
        return str(uuid.uuid5(NS_V0, f"lexeme|{COURSE_SLUG}|{lemma}|{pos}"))
    return uid("lexeme", lemma, pos)


def row(values: list) -> str:
    return "(" + ", ".join(q(v) for v in values) + ")"


def batched(table: str, columns: list[str], rows: list[list], conflict: str) -> list[str]:
    """Instrucoes `INSERT ... VALUES (...),(...)`: uma linha de dados por linha do arquivo (o goose divide por `;` no fim da linha)."""
    out = []
    for start in range(0, len(rows), BATCH):
        chunk = rows[start:start + BATCH]
        lines = [f"INSERT INTO {table} ({', '.join(columns)}) VALUES"]
        lines += [row(r) + ("," if i < len(chunk) - 1 else "") for i, r in enumerate(chunk)]
        lines.append(f"ON CONFLICT ({conflict}) DO NOTHING;")
        out.append("\n".join(lines))
    return out


def source_ids() -> dict[str, str]:
    return {code: uid("source", code) for code in ("ngsl", "tatoeba")} | {"original": str(uuid.uuid5(NS_V0, "source|wordloop-original"))}


def base_sql(units: list[dict], words: list[dict]) -> str:
    sid = source_ids()
    manifest = json.loads((V1 / "sources.json").read_text(encoding="utf-8"))
    retrieved = manifest["retrieved_at"]
    sources = [
        [sid["ngsl"], "ngsl", "New General Service List 1.2", "https://www.newgeneralservicelist.com/", "1.2", "CC-BY-SA-4.0",
         "New General Service List (NGSL) 1.2, de Browne, Culligan e Phillips (newgeneralservicelist.com), licenciada sob CC BY-SA 4.0. "
         "Lista de frequência usada para escolher e ordenar o vocabulário.", retrieved, "Lista de frequência; sem texto copiado além dos lemas."],
        [sid["tatoeba"], "tatoeba", "Tatoeba", "https://tatoeba.org/", f"exportação de {retrieved}", "CC-BY-2.0",
         "Frases em inglês e traduções em português do projeto Tatoeba (tatoeba.org), escritas por seus colaboradores e licenciadas sob "
         "CC BY 2.0 FR. Cada frase guarda o número original e o autor.", retrieved,
         "Licença CC BY 2.0 FR (versão francesa); incorporada com crédito. A compatibilidade com CC BY-SA 4.0 deve ser conferida por revisão jurídica."],
    ]
    out = [HEADER + "-- +goose Up"]
    out += batched("source", ["id", "code", "name", "url", "version", "license_spdx", "attribution_notice", "retrieved_at", "notes"], sources, "code")
    out.append("UPDATE source SET attribution_notice = " + q(
        "Textos originais do projeto wordloop, licenciados sob CC BY-SA 4.0. As traduções e glosas em pt-BR são assistidas por IA e "
        "não passaram por revisão humana: encontrou um erro? Avise-nos.") + " WHERE code = 'wordloop-original';")
    unit_rows = [[uid("unit", COURSE_SLUG, str(u["position"])), COURSE_ID, u["position"], u["theme_key"]] for u in units]
    out += batched("unit", ["id", "course_id", "position", "theme_key"], unit_rows, "course_id, position")
    lex_rows = [[lexeme_id(w["lemma"], w["pos"]), COURSE_ID, w["lemma"], w["pos"], w["rank"]] for w in words
                if (w["lemma"], w["pos"]) not in V0_LEXEMES]
    out += batched("lexeme", ["id", "course_id", "lemma", "pos", "freq_rank"], lex_rows, "course_id, lemma, pos")
    return "\n".join(out) + "\n\n" + DOWN


def origin(entry: dict) -> tuple[str, str]:
    return f"eng#{entry['eng_id']}/por#{entry['por_id']}", f"{entry['eng_user']} / {entry['por_user']}"


def content_sql(items: list[list], texts: list[list], attributions: list[list]) -> str:
    item_cols = ["id", "course_id", "unit_id", "lexeme_id", "kind", "natural_key", "sense_no", "cefr", "concreteness", "status", "content_version"]
    text_cols = ["item_id", "role", "locale", "position", "body", "source_id", "source_ref"]
    attr_cols = ["id", "entity_type", "entity_id", "source_id", "creator", "source_ref", "notice"]
    out = [HEADER + "-- +goose Up"]
    out += batched("item", item_cols, items, "course_id, natural_key")
    out += batched("item_text", text_cols, texts, "item_id, role, locale, position")
    out += batched("attribution", attr_cols, attributions, "id")
    return "\n".join(out) + "\n\n" + DOWN


NOTICE_TATOEBA = "Tatoeba (tatoeba.org), CC BY 2.0 FR."


def words_sql(units: list[dict], words: list[dict]) -> str:
    sid, unit_id = source_ids(), {u["position"]: uid("unit", COURSE_SLUG, str(u["position"])) for u in units}
    items, texts, attributions = [], [], []
    for w in words:
        key = f"word|{w['lemma']}|{w['pos']}|1"
        iid = uid("item", key)
        items.append([iid, COURSE_ID, unit_id[w["unit"]], lexeme_id(w["lemma"], w["pos"]), "word", key, 1, w["cefr"],
                      "function" if w["pos"] in FUNCTION_POS else None, "published", 1])
        texts.append([iid, "term", "en-US", 0, w["lemma"], sid["ngsl"], f"rank {w['rank']}"])
        texts.append([iid, "gloss", "pt-BR", 0, w["gloss"], sid["original"], None])
        if "ex" in w:
            ref, creators = origin(w["ex"])
            texts.append([iid, "example", "en-US", 0, w["ex"]["text"], sid["tatoeba"], ref])
            texts.append([iid, "example_gloss", "pt-BR", 0, w["ex"]["pt"], sid["tatoeba"], ref])
            attributions.append([uid("attribution", iid), "item", iid, sid["tatoeba"], creators, ref, NOTICE_TATOEBA])
    return content_sql(items, texts, attributions)


def sentences_sql(units: list[dict], sentences: list[dict]) -> str:
    sid, unit_id = source_ids(), {u["position"]: uid("unit", COURSE_SLUG, str(u["position"])) for u in units}
    items, texts, attributions = [], [], []
    for s in sentences:
        key = f"sentence|{norm(s['text'])}"
        iid = uid("item", key)
        ref, creators = origin(s)
        items.append([iid, COURSE_ID, unit_id[s["unit"]], None, "sentence", key, None, s["cefr"], None, "published", 1])
        texts.append([iid, "term", "en-US", 0, s["text"], sid["tatoeba"], ref])
        texts.append([iid, "gloss", "pt-BR", 0, s["pt"], sid["tatoeba"], ref])
        attributions.append([uid("attribution", iid), "item", iid, sid["tatoeba"], creators, ref, NOTICE_TATOEBA])
    return content_sql(items, texts, attributions)


def publish_sql() -> str:
    return (HEADER + "-- +goose Up\n-- O seed v0 nasceu como rascunho e nenhum caminho o publicava: sem isto, quem não é admin não vê nem estuda nada.\n"
            "-- Os 125 itens passaram na conferência do verify-v0.sql; as glosas continuam marcadas como assistidas por IA nos créditos.\n"
            f"UPDATE item SET status = 'published', updated_at = now() WHERE course_id = {q(COURSE_ID)} AND status = 'draft';\n\n" + DOWN)


def main() -> int:
    units, words, sentences = load()
    problems = validate(units, words, sentences)
    v0_sentences = {norm(s["text"]) for s in SEED_V0["sentences"]}
    problems += [f"frase ja existe no v0: {s['text']!r}" for s in sentences if norm(s["text"]) in v0_sentences]
    if problems:
        print("\n".join(f"ERRO {p}" for p in problems[:30]))
        return 1
    files = {
        "00020_content_v1_base.sql": base_sql(units, words),
        "00021_content_v1_words.sql": words_sql(units, words),
        "00022_content_v1_sentences.sql": sentences_sql(units, sentences),
        "00023_content_publish.sql": publish_sql(),
    }
    for name, text in files.items():
        (MIGRATIONS / name).write_text(text, encoding="utf-8", newline="")
        print(f"{name}: {len(text.splitlines())} linhas, {len(text) // 1024} KB")
    return 0


if __name__ == "__main__":
    sys.exit(main())
