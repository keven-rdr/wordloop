"""Confere as migracoes geradas (api/migrations/00020..00023): atualizadas em relacao ao JSON e integras como o Postgres veria.

Sem banco aqui, o teste le o SQL com um analisador pequeno (aspas simples com '' escapado) e confere o que o Postgres exigiria:
colunas, chaves unicas, chaves estrangeiras e a divisao de instrucoes do goose (so `;` no fim da linha encerra uma instrucao).
"""
import re
import unittest

import gen_v1
from common import MIGRATIONS
from validate import load

FILES = ["00020_content_v1_base.sql", "00021_content_v1_words.sql", "00022_content_v1_sentences.sql", "00023_content_publish.sql"]
HEADER = re.compile(r"^INSERT INTO (\w+) \(([^)]*)\) VALUES$")


def parse_values(line: str) -> list:
    """`('a', 1, NULL, 'it''s'),` -> ['a', 1, None, "it's"]."""
    text = line.rstrip(",")
    assert text.startswith("(") and text.endswith(")"), line[:80]
    out, i, n = [], 1, len(text) - 1
    while i < n:
        c = text[i]
        if c == "'":
            j, buf = i + 1, []
            while True:
                if text[j] == "'" and j + 1 < len(text) and text[j + 1] == "'":
                    buf.append("'")
                    j += 2
                elif text[j] == "'":
                    break
                else:
                    buf.append(text[j])
                    j += 1
            out.append("".join(buf))
            i = j + 1
        elif text.startswith("NULL", i):
            out.append(None)
            i += 4
        elif c.isdigit() or c == "-":
            j = i
            while j < n and text[j] not in ", ":
                j += 1
            out.append(int(text[i:j]))
            i = j
        else:
            i += 1  # virgula ou espaco entre valores
    return out


def read_inserts(name: str) -> dict[str, list[dict]]:
    tables: dict[str, list[dict]] = {}
    lines = (MIGRATIONS / name).read_text(encoding="utf-8").splitlines()
    i = 0
    while i < len(lines):
        match = HEADER.match(lines[i])
        if not match:
            i += 1
            continue
        table, columns = match.group(1), [c.strip() for c in match.group(2).split(",")]
        i += 1
        while not lines[i].startswith("ON CONFLICT"):
            values = parse_values(lines[i])
            assert len(values) == len(columns), f"{name}:{i + 1}: {len(values)} valores para {len(columns)} colunas"
            tables.setdefault(table, []).append(dict(zip(columns, values, strict=True)))
            i += 1
        assert lines[i].endswith("DO NOTHING;"), f"{name}:{i + 1}: instrucao sem `;` final"
        i += 1
    return tables


class MigrationsTest(unittest.TestCase):
    @classmethod
    def setUpClass(cls):
        cls.units, cls.words, cls.sentences = load()
        cls.base = read_inserts(FILES[0])
        cls.words_sql = read_inserts(FILES[1])
        cls.sent_sql = read_inserts(FILES[2])

    def test_files_are_up_to_date_with_the_json(self):
        expected = {
            FILES[0]: gen_v1.base_sql(self.units, self.words),
            FILES[1]: gen_v1.words_sql(self.units, self.words),
            FILES[2]: gen_v1.sentences_sql(self.units, self.sentences),
            FILES[3]: gen_v1.publish_sql(),
        }
        for name, text in expected.items():
            self.assertEqual((MIGRATIONS / name).read_text(encoding="utf-8"), text, f"{name} desatualizado: rode content/tools/gen_v1.py")

    def test_goose_splits_only_at_the_end_of_each_statement(self):
        for name in FILES:
            statements = 0
            for line in (MIGRATIONS / name).read_text(encoding="utf-8").splitlines():
                if line.startswith("--") or not line.strip():
                    continue
                if line.rstrip().endswith(";"):
                    statements += 1
                    self.assertTrue(line.startswith(("ON CONFLICT", "UPDATE", "SELECT")), f"{name}: `;` fora do fim de instrucao: {line[:60]}")
            self.assertGreater(statements, 0, name)

    def test_row_counts_and_batch_size(self):
        self.assertEqual(len(self.base["unit"]), len(self.units))
        self.assertEqual(len(self.words_sql["item"]), len(self.words))
        self.assertEqual(len(self.sent_sql["item"]), len(self.sentences))
        statements = sum(len(read_inserts(f).get("item_text", [])) // gen_v1.BATCH + 1 for f in FILES[1:3])
        self.assertLess(statements, 80, "a carga precisa ser poucas dezenas de instrucoes (round trips ao Neon)")

    def test_referential_integrity_as_postgres_would_check(self):
        v0_lexemes = {gen_v1.lexeme_id(l, p) for l, p in gen_v1.V0_LEXEMES}
        lexemes = {r["id"] for r in self.base["lexeme"]} | v0_lexemes
        units = {r["id"] for r in self.base["unit"]}
        sources = {r["id"] for r in self.base["source"]} | {gen_v1.source_ids()["original"]}
        for sql in (self.words_sql, self.sent_sql):
            items = {r["id"] for r in sql["item"]}
            self.assertEqual(len(items), len(sql["item"]), "id de item repetido")
            self.assertEqual(len({r["natural_key"] for r in sql["item"]}), len(sql["item"]), "chave natural repetida")
            for item in sql["item"]:
                self.assertIn(item["unit_id"], units)
                self.assertTrue(item["lexeme_id"] is None or item["lexeme_id"] in lexemes)
                self.assertEqual(item["status"], "published")
                self.assertEqual(item["course_id"], gen_v1.COURSE_ID)
            keys = [(t["item_id"], t["role"], t["locale"], t["position"]) for t in sql["item_text"]]
            self.assertEqual(len(set(keys)), len(keys), "item_text repetido")
            for text in sql["item_text"]:
                self.assertIn(text["item_id"], items)
                self.assertIn(text["source_id"], sources)
                self.assertTrue(text["body"].strip())
            for attribution in sql["attribution"]:
                self.assertIn(attribution["entity_id"], items)
                self.assertIn(attribution["source_id"], sources)

    def test_every_item_has_term_and_pt_gloss(self):
        for sql in (self.words_sql, self.sent_sql):
            roles: dict[str, set] = {}
            for t in sql["item_text"]:
                roles.setdefault(t["item_id"], set()).add((t["role"], t["locale"]))
            for item in sql["item"]:
                self.assertLessEqual({("term", "en-US"), ("gloss", "pt-BR")}, roles[item["id"]], item["natural_key"])

    def test_v0_lexemes_are_reused_not_duplicated(self):
        pairs = [(r["lemma"], r["pos"]) for r in self.base["lexeme"]]
        self.assertEqual(len(set(pairs)), len(pairs))
        self.assertFalse(set(pairs) & gen_v1.V0_LEXEMES)


if __name__ == "__main__":
    unittest.main()
