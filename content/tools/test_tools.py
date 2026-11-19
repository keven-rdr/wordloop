"""Testes das ferramentas de conteudo (unittest, sem rede e sem as fontes baixadas): python -m unittest discover content/tools"""
import copy
import unittest

import common
import ngsl
import pick
import tatoeba
import validate


def fake_ngsl() -> ngsl.Ngsl:
    ranks = {"the": 1, "be": 2, "i": 3, "go": 4, "house": 5, "red": 6, "walk": 7, "can": 8, "not": 9, "rare": 5000}
    forms = {"the": "the", "be": "be", "am": "be", "is": "be", "was": "be", "i": "i", "go": "go", "went": "go", "house": "house",
             "red": "red", "walk": "walk", "walked": "walk", "can": "can", "don": "do", "do": "do", "not": "not", "rare": "rare"}
    ranks["do"] = 10
    return ngsl.Ngsl(ranks, forms)


class NgslTest(unittest.TestCase):
    def test_lemma_of_handles_forms_contractions_and_names(self):
        n = fake_ngsl()
        self.assertEqual(n.lemma_of("went"), "go")
        self.assertEqual(n.lemma_of("don't"), "do")
        self.assertEqual(n.lemma_of("i'm"), "i")
        self.assertEqual(n.lemma_of("tom"), "tom")
        self.assertEqual(n.lemma_of("tom's"), "tom")
        self.assertIsNone(n.lemma_of("zebra"))
        self.assertIsNone(n.lemma_of("house'zz"))

    def test_required_rank_is_the_rarest_word_and_unknown_words_reject(self):
        n = fake_ngsl()
        self.assertIsNone(n.required_rank(["i", "went", "home"]), "home nao e do vocabulario")
        self.assertEqual(n.required_rank(["i", "went"]), 4)
        self.assertEqual(n.required_rank(["tom", "is", "red"]), 6)
        self.assertIsNone(n.required_rank(["i", "rare"]), "passa do vocabulario do lote")
        self.assertIsNone(n.required_rank(["tom"]), "so nomes: nada a ensinar")

    def test_lessons_and_levels(self):
        self.assertEqual(ngsl.lesson_index(1), 0)
        self.assertEqual(ngsl.lesson_index(30), 0)
        self.assertEqual(ngsl.lesson_index(31), 1)
        self.assertEqual([ngsl.cefr_of(r) for r in (1, 600, 601, 1500, 1501)], ["A1", "A1", "A2", "A2", "B1"])


class TatoebaFilterTest(unittest.TestCase):
    def test_english_side(self):
        ok = tatoeba._eng_ok
        self.assertTrue(ok("I like the red house."))
        for bad in ("I have 3 cats.", "he runs fast.", "Is it OK?!!", "Hello.", "I like it, \"yes\".", "I like cats. You too.",
                    "The war is over now.", "It is a very very very very very long one."):
            self.assertFalse(ok(bad), bad)

    def test_portuguese_side(self):
        ok = tatoeba._por_ok
        self.assertTrue(ok("Eu gosto da casa vermelha.", "I like the red house."))
        self.assertFalse(ok("Eu estou a ler um livro.", "I am reading a book."), "pt-PT")
        self.assertFalse(ok("Eu gosto muito da casa vermelha", "I like the red house."), "sem pontuacao final")
        self.assertFalse(ok("Sim.", "I like the big red house over there."), "traducao curta demais")
        self.assertFalse(ok("O Tom gosta da casa.", "She likes the house."), "inventou o Tom")
        self.assertFalse(ok("Ela gosta da casa.", "Tom likes the house."), "sumiu o Tom")
        self.assertTrue(ok("Maria gosta da casa.", "Mary likes the house."))


def cand(eng_id: int, text: str, rank: int, lemmas=()) -> pick.Candidate:
    pair = tatoeba.Pair(eng_id, text, "u", eng_id + 1000, "pt " + text, "v")
    return pick.Candidate(pair, rank, frozenset(lemmas))


class PickTest(unittest.TestCase):
    def test_sentences_go_to_the_lesson_of_their_rarest_word_with_a_cap(self):
        n = fake_ngsl()
        cands = [cand(i, "I walk a lot here now", 7 if i < 5 else 45, ["walk"]) for i in range(10)]
        chosen = pick.pick_sentences(cands, n, per_lesson=3)
        self.assertEqual(len(chosen), 6)
        self.assertEqual({c.lesson for c in chosen}, {0, 1})

    def test_examples_prefer_unused_short_sentences_with_the_exact_word(self):
        a = cand(1, "I walk home today now", 7, ["walk", "i"])
        b = cand(2, "I walked home today now", 7, ["walk", "i"])
        c = cand(3, "Walk with me now please", 7, ["walk"])
        got = pick.pick_examples([(7, "walk")], [a, b, c], taken={1})
        self.assertEqual(got["walk"].pair.eng_id, 3, "a 1 ja virou frase; a 2 nao tem a forma exata")

    def test_examples_never_use_rarer_words_than_the_word_itself(self):
        hard = cand(1, "I walk home today now", 900, ["walk"])
        self.assertNotIn("walk", pick.pick_examples([(7, "walk")], [hard], set()))


GOOD_UNITS = [{"position": 7, "theme_key": "lesson-a1-01"}]


def good_data():
    words = [{"rank": i, "lemma": f"w{i}", "pos": "noun", "gloss": "palavra", "cefr": "A1", "unit": 7} for i in range(2000)]
    sentences = [{"text": f"I like the red house number {chr(97 + i % 26)}{chr(97 + i // 26 % 26)}{chr(97 + i // 676)}.", "pt": "Eu gosto da casa.", "cefr": "A1",
                  "unit": 7, "rank": 5, "eng_id": i + 1, "por_id": i + 9, "eng_user": "a", "por_user": "b"} for i in range(2000)]
    return copy.deepcopy(GOOD_UNITS), words, sentences


class ValidateTest(unittest.TestCase):
    def test_good_data_passes(self):
        units, words, sentences = good_data()
        self.assertEqual(validate.sentence_problems(sentences[0], {7}), [])
        self.assertEqual(validate.validate(units, words, sentences), [])

    def test_each_problem_is_reported(self):
        units, words, sentences = good_data()
        cases = {
            "pos": lambda w, s: w[0].update(pos="gerund"),
            "glosa": lambda w, s: w[1].update(gloss=" "),
            "unidade": lambda w, s: w[2].update(unit=99),
            "repetida": lambda w, s: w[3].update(lemma=w[4]["lemma"]),
            "pecas": lambda w, s: s[0].update(text="Go."),
            "bloqueado": lambda w, s: s[1].update(text="I will kill the red house."),
            "Portugal": lambda w, s: s[2].update(pt="Eu estou a ler."),
            "origem": lambda w, s: s[3].update(eng_id=0),
            "frase repetida": lambda w, s: s[5].update(text=s[4]["text"]),
            "poucas palavras": lambda w, s: w.pop(),
        }
        for name, mutate in cases.items():
            w, s = copy.deepcopy(words), copy.deepcopy(sentences)
            mutate(w, s)
            self.assertTrue(validate.validate(units, w, s), name)


class CommonTest(unittest.TestCase):
    def test_norm_and_q(self):
        self.assertEqual(common.norm("  Don't   Go, Tom! "), "don't go tom")
        self.assertEqual(common.q("it's"), "'it''s'")
        self.assertEqual(common.q(None), "NULL")
        self.assertEqual(common.q(3), "3")


if __name__ == "__main__":
    unittest.main()
