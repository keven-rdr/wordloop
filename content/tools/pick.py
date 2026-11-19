"""Escolha das frases do lote: cada frase entra na licao da palavra mais rara que ela usa, e cada palavra ganha um exemplo."""
from dataclasses import dataclass

from ngsl import LESSON_WORDS, Ngsl, lesson_index
from tatoeba import Pair

PER_LESSON = 30  # frases por licao (teto)
IDEAL_LENGTHS = (4, 5, 6, 7, 8)  # palavras: varia por frase (id) para nao ficar tudo do mesmo tamanho


@dataclass(frozen=True)
class Candidate:
    pair: Pair
    rank: int  # posicao da palavra mais rara
    lemmas: frozenset[str]

    @property
    def lesson(self) -> int:
        return lesson_index(self.rank)


def candidates(pairs: list[Pair], ngsl: Ngsl) -> list[Candidate]:
    out = []
    for pair in pairs:
        words = pair.words
        rank = ngsl.required_rank(words)
        if rank is not None:
            out.append(Candidate(pair, rank, frozenset(ngsl.lemma_of(w) or w for w in words)))
    return out


def _lesson_words(lesson: int) -> range:
    return range(lesson * LESSON_WORDS + 1, (lesson + 1) * LESSON_WORDS + 1)


def pick_sentences(cands: list[Candidate], ngsl: Ngsl, per_lesson: int = PER_LESSON) -> list[Candidate]:
    """Ate `per_lesson` frases por licao, priorizando as que usam as palavras novas daquela licao."""
    by_lesson: dict[int, list[Candidate]] = {}
    for c in cands:
        by_lesson.setdefault(c.lesson, []).append(c)
    chosen = []
    for lesson in sorted(by_lesson):
        new = {lemma for lemma, rank in ngsl.rank.items() if rank in _lesson_words(lesson)}

        def score(c: Candidate, new=new):
            return (-len(c.lemmas & new), abs(len(c.pair.words) - IDEAL_LENGTHS[c.pair.eng_id % len(IDEAL_LENGTHS)]), c.pair.eng_id)

        chosen.extend(sorted(by_lesson[lesson], key=score)[:per_lesson])
    return chosen


def pick_examples(words: list[tuple[int, str]], pool: list[Candidate], taken: set[int]) -> dict[str, Candidate]:
    """Um exemplo curto por palavra, usando so palavras tao frequentes quanto ela. Prefere frases ainda nao usadas."""
    index: dict[str, list[Candidate]] = {}
    for c in pool:
        for lemma in c.lemmas:
            index.setdefault(lemma, []).append(c)
    used = set(taken)
    examples: dict[str, Candidate] = {}
    for rank, lemma in words:
        options = [c for c in index.get(lemma, []) if c.rank <= max(rank, 60)]
        exact = [c for c in options if lemma in c.pair.words]

        def score(c: Candidate):
            return (c.pair.eng_id in used, abs(len(c.pair.words) - 5), c.pair.eng_id)

        best = sorted(exact or options, key=score)
        if best:
            examples[lemma] = best[0]
            used.add(best[0].pair.eng_id)
    return examples
