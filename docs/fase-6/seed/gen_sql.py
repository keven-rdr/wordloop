#!/usr/bin/env python3
"""Gerador do seed (Fase 6): fonte JSON -> SQL idempotente + SQL de conferência.

Uso:  python gen_sql.py [--source source/sample-seed.json] [--out out]

Princípios (mesmos do `seed-referencias` do Contratos, escritos do zero):
  * Reproduzível: mesmo JSON => mesmo SQL (UUIDs v5 derivados da chave natural, ordem fixa).
  * Idempotente: INSERT ... ON CONFLICT na chave natural; rodar duas vezes não duplica nada.
  * Conferível: um segundo SQL reconcilia o banco com a fonte (contagens + violações).
  * Não sobrescreve decisões humanas: `status` e `content_version` NÃO entram no DO UPDATE.
"""
import argparse
import json
import pathlib
import unicodedata
import uuid


def norm(s: str) -> str:
    return unicodedata.normalize("NFKC", s).strip().lower()


def q(v) -> str:
    if v is None:
        return "NULL"
    if isinstance(v, (int, float)):
        return str(v)
    return "'" + str(v).replace("'", "''") + "'"


class Gen:
    def __init__(self, data: dict):
        self.d = data
        self.ns = uuid.uuid5(uuid.NAMESPACE_URL, data["namespace"])
        self.sql: list[str] = []
        self.counts = {"word": 0, "expression": 0, "phrasal_verb": 0, "sentence": 0}

    def uid(self, *parts: str) -> str:
        return str(uuid.uuid5(self.ns, "|".join(parts)))

    def emit(self, text: str) -> None:
        self.sql.append(text)

    def insert(self, table: str, row: dict, conflict: str, update_cols: list[str]) -> None:
        cols = ", ".join(row)
        vals = ", ".join(q(v) if not (isinstance(v, str) and v == "now()") else "now()" for v in row.values())
        if update_cols:
            sets = ", ".join(f"{c} = EXCLUDED.{c}" for c in update_cols)
            tail = f"ON CONFLICT ({conflict}) DO UPDATE SET {sets}"
        else:
            tail = f"ON CONFLICT ({conflict}) DO NOTHING"
        self.emit(f"INSERT INTO {table} ({cols}) VALUES ({vals}) {tail};")

    def build(self) -> str:
        d = self.d
        self.emit("-- GERADO por gen_sql.py a partir de source/sample-seed.json. NÃO EDITE À MÃO.")
        self.emit("-- Idempotente: pode rodar várias vezes. Depois rode sample-verify.sql.")
        self.emit("BEGIN;")

        for s in d["sources"]:
            row = {"id": self.uid("source", s["code"]), "code": s["code"], "name": s["name"], "url": s.get("url"),
                   "version": s.get("version"), "license_spdx": s["license_spdx"],
                   "attribution_notice": s["attribution_notice"], "notes": s.get("notes")}
            self.insert("source", row, "code", ["name", "url", "version", "license_spdx", "attribution_notice", "notes"])
        src_orig = self.uid("source", "wordloop-original")
        src_icon = self.uid("source", "openmoji")

        sub = d["subject"]
        subject_id = self.uid("subject", sub["code"])
        self.insert("subject", {"id": subject_id, "code": sub["code"], "name_key": sub["name_key"]}, "code", ["name_key"])

        c = d["course"]
        course_id = self.uid("course", c["slug"])
        self.insert("course", {"id": course_id, "subject_id": subject_id, "slug": c["slug"],
                               "studied_locale": c["studied_locale"], "help_locale": c["help_locale"], "status": c["status"]},
                    "slug", ["studied_locale", "help_locale"])
        self.course_id = course_id

        unit_ids = {}
        for u in d["units"]:
            unit_ids[u["position"]] = self.uid("unit", c["slug"], str(u["position"]))
            self.insert("unit", {"id": unit_ids[u["position"]], "course_id": course_id, "position": u["position"],
                                 "theme_key": u["theme_key"]}, "course_id, position", ["theme_key"])

        lexemes = {}
        for w in d["words"]:
            lexemes[(norm(w["lemma"]), w["pos"])] = None
        for e in d["lexemes_extra"]:
            lexemes[(norm(e["lemma"]), e["pos"])] = None
        for (lemma, pos) in sorted(lexemes):
            lexemes[(lemma, pos)] = self.uid("lexeme", c["slug"], lemma, pos)
            self.insert("lexeme", {"id": lexemes[(lemma, pos)], "course_id": course_id, "lemma": lemma, "pos": pos},
                        "course_id, lemma, pos", [])

        self.item_ids: dict[str, str] = {}

        def add_item(kind, key, unit, lexeme, sense_no, cefr, conc):
            iid = self.uid("item", c["slug"], key)
            self.item_ids[key] = iid
            self.counts[kind] += 1
            self.insert("item", {"id": iid, "course_id": course_id, "unit_id": unit, "lexeme_id": lexeme, "kind": kind,
                                 "natural_key": key, "sense_no": sense_no, "cefr": cefr, "concreteness": conc,
                                 "status": "draft", "content_version": 1},
                        "course_id, natural_key", ["unit_id", "lexeme_id", "sense_no", "cefr", "concreteness", "updated_at"])
            return iid

        def add_text(iid, role, locale, body, pos=0):
            self.insert("item_text", {"item_id": iid, "role": role, "locale": locale, "position": pos, "body": body,
                                      "source_id": src_orig}, "item_id, role, locale, position", ["body", "source_id"])

        for w in d["words"]:
            lemma, sense = norm(w["lemma"]), w.get("sense_no", 1)
            key = f"word|{lemma}|{w['pos']}|{sense}"
            iid = add_item("word", key, unit_ids[w["unit"]], lexemes[(lemma, w["pos"])], sense, w["cefr"], w["conc"])
            add_text(iid, "term", "en-US", w["lemma"])
            add_text(iid, "gloss", "pt-BR", w["gloss"])
            add_text(iid, "example", "en-US", w["ex"])
            add_text(iid, "example_gloss", "pt-BR", w["ex_pt"])

        for e in d["expressions"]:
            key = f"expression|{norm(e['text'])}|1"
            iid = add_item("expression", key, None, None, 1, e["cefr"], "function")
            add_text(iid, "term", "en-US", e["text"])
            add_text(iid, "gloss", "pt-BR", e["gloss"])
            add_text(iid, "example", "en-US", e["ex"])
            add_text(iid, "example_gloss", "pt-BR", e["ex_pt"])

        for p in d["phrasal_verbs"]:
            key = f"phrasal_verb|{norm(p['text'])}|1"
            iid = add_item("phrasal_verb", key, None, lexemes[(norm(p["base"]), "verb")], 1, p["cefr"], "abstract")
            add_text(iid, "term", "en-US", p["text"])
            add_text(iid, "gloss", "pt-BR", p["gloss"])
            add_text(iid, "example", "en-US", p["ex"])
            add_text(iid, "example_gloss", "pt-BR", p["ex_pt"])

        for s in d["sentences"]:
            key = f"sentence|{norm(s['text'])}"
            iid = add_item("sentence", key, None, None, None, "A1", "function")
            add_text(iid, "term", "en-US", s["text"])
            add_text(iid, "gloss", "pt-BR", s["pt"])
            cloze = s["text"].replace(s["target"], "{{" + s["target"] + "}}", 1)
            assert cloze != s["text"], f"alvo não encontrado na frase: {s}"
            add_text(iid, "cloze", "en-US", cloze)

        self.n_relations = 0
        for r in d["relations"]:
            self.n_relations += 1
            self.insert("item_relation", {"from_item_id": self.item_ids[r["from"]], "to_item_id": self.item_ids[r["to"]],
                                          "type": r["type"], "weight": r["weight"], "min_stability_days": 21,
                                          "source_id": src_orig}, "from_item_id, to_item_id, type", ["weight", "min_stability_days"])

        self.n_media = 0
        for m in d["media"]:
            self.n_media += 1
            mid = self.uid("media", "openmoji", m["hexcode"])
            notice = self.d["sources"][1]["attribution_notice"]
            self.insert("media", {"id": mid, "kind": "image", "storage_key": f"openmoji/{m['hexcode']}.svg",
                                  "sha256": "pending:calcular-na-ingestao", "mime": "image/svg+xml", "source_id": src_icon,
                                  "origin_url": "https://openmoji.org/", "creator": "OpenMoji contributors",
                                  "license_spdx": "CC-BY-SA-4.0", "attribution_text": notice, "status": "pending"},
                        "storage_key", ["license_spdx", "attribution_text"])
            self.insert("item_media", {"item_id": self.item_ids[m["item"]], "media_id": mid, "role": "context", "position": 0},
                        "item_id, media_id, role", [])
            self.insert("attribution", {"id": self.uid("attribution", "media", mid), "entity_type": "media", "entity_id": mid,
                                        "source_id": src_icon, "creator": "OpenMoji contributors", "source_ref": m["hexcode"],
                                        "notice": notice}, "id", ["notice"])

        self.emit("COMMIT;")
        return "\n".join(self.sql) + "\n"

    def verify(self) -> str:
        cid = self.course_id
        n_items = sum(self.counts.values())
        n_texts = (self.counts["word"] + self.counts["expression"] + self.counts["phrasal_verb"]) * 4 + self.counts["sentence"] * 3
        out = ["-- GERADO por gen_sql.py. Reconcilia o banco com a fonte.",
               "-- Bloco 1: contagens (status deve ser 'ok'). Bloco 2: violações (devem retornar 0 linhas)."]
        for label, sql_from, expected in [
            ("item.word", f"item WHERE course_id = '{cid}' AND kind = 'word'", self.counts["word"]),
            ("item.expression", f"item WHERE course_id = '{cid}' AND kind = 'expression'", self.counts["expression"]),
            ("item.phrasal_verb", f"item WHERE course_id = '{cid}' AND kind = 'phrasal_verb'", self.counts["phrasal_verb"]),
            ("item.sentence", f"item WHERE course_id = '{cid}' AND kind = 'sentence'", self.counts["sentence"]),
            ("item.total", f"item WHERE course_id = '{cid}'", n_items),
            ("item_text.total", f"item_text WHERE item_id IN (SELECT id FROM item WHERE course_id = '{cid}')", n_texts),
            ("item_relation.total", f"item_relation WHERE from_item_id IN (SELECT id FROM item WHERE course_id = '{cid}')", self.n_relations),
            ("media.total", "media WHERE source_id IN (SELECT id FROM source WHERE code = 'openmoji')", self.n_media),
        ]:
            out.append(f"SELECT '{label}' AS check_name, COUNT(*) AS found, {expected} AS expected, "
                       f"CASE WHEN COUNT(*) = {expected} THEN 'ok' ELSE 'FALHA' END AS status FROM {sql_from};")
        v = [
            ("item sem gloss pt-BR",
             f"SELECT i.natural_key FROM item i WHERE i.course_id = '{cid}' AND NOT EXISTS "
             "(SELECT 1 FROM item_text t WHERE t.item_id = i.id AND t.role = 'gloss' AND t.locale = 'pt-BR')"),
            ("item publicado sem revisão (todo item da amostra deve ser draft)",
             f"SELECT natural_key FROM item WHERE course_id = '{cid}' AND status <> 'draft'"),
            ("cloze sem exatamente uma lacuna",
             "SELECT item_id FROM item_text WHERE role = 'cloze' AND "
             "(LENGTH(body) - LENGTH(REPLACE(body, '{{', ''))) <> 2"),
            ("item_text sem fonte",
             f"SELECT item_id FROM item_text WHERE source_id IS NULL AND item_id IN (SELECT id FROM item WHERE course_id = '{cid}')"),
            ("palavra sem sentido 1 no seu lema",
             f"SELECT l.lemma, l.pos FROM lexeme l WHERE l.course_id = '{cid}' AND EXISTS (SELECT 1 FROM item i WHERE i.lexeme_id = l.id AND i.kind = 'word') "
             "AND NOT EXISTS (SELECT 1 FROM item i WHERE i.lexeme_id = l.id AND i.kind = 'word' AND i.sense_no = 1)"),
            ("mídia sem licença ou crédito",
             "SELECT storage_key FROM media WHERE license_spdx IS NULL OR attribution_text IS NULL OR attribution_text = ''"),
            ("mídia aprovada sem hash real (deve estar pending até a ingestão)",
             "SELECT storage_key FROM media WHERE status = 'approved' AND sha256 LIKE 'pending:%'"),
            ("relação consigo mesma ou fora do curso",
             f"SELECT r.from_item_id FROM item_relation r WHERE r.from_item_id = r.to_item_id OR r.to_item_id NOT IN (SELECT id FROM item WHERE course_id = '{cid}')"),
            ("item de palavra com mais de um primeiro sentido para o mesmo lema",
             f"SELECT lexeme_id FROM item WHERE course_id = '{cid}' AND kind = 'word' AND sense_no = 1 GROUP BY lexeme_id HAVING COUNT(*) > 1"),
        ]
        for name, sql in v:
            out.append(f"-- violação: {name}\nSELECT '{name}' AS violation, x.* FROM ({sql}) AS x;")
        return "\n".join(out) + "\n"


def main() -> None:
    ap = argparse.ArgumentParser()
    ap.add_argument("--source", default="source/sample-seed.json")
    ap.add_argument("--out", default="out")
    a = ap.parse_args()
    data = json.loads(pathlib.Path(a.source).read_text(encoding="utf-8"))
    g = Gen(data)
    seed = g.build()
    out = pathlib.Path(a.out)
    out.mkdir(parents=True, exist_ok=True)
    (out / "sample-seed.sql").write_text(seed, encoding="utf-8")
    (out / "sample-verify.sql").write_text(g.verify(), encoding="utf-8")
    print(f"ok: {sum(g.counts.values())} itens {g.counts}, {g.n_relations} relações, {g.n_media} mídias -> {out}/")


if __name__ == "__main__":
    main()
