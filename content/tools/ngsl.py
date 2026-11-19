"""NGSL 1.2: ranking de lemas e mapa forma -> lema (para saber se uma palavra de uma frase e conhecida)."""
import csv

from common import CACHE, NAMES

LESSON_WORDS = 30  # palavras por licao (ranking de frequencia)
MAX_RANK = 2100  # vocabulario do lote v1: os 2.100 lemas mais frequentes (menos os que ja estao no v0)
CONTRACTION_SUFFIXES = {"s", "t", "m", "re", "ve", "ll", "d"}


class Ngsl:
    def __init__(self, ranks: dict[str, int], forms: dict[str, str]):
        self.rank = ranks  # lema -> posicao (1 = mais frequente)
        self.forms = forms  # forma -> lema

    @classmethod
    def load(cls) -> "Ngsl":
        ranks: dict[str, int] = {}
        with open(CACHE / "ngsl_stats.csv", encoding="utf-8", newline="") as f:
            for position, row in enumerate(csv.DictReader(f), start=1):
                ranks.setdefault(row["Lemma"].strip().lower(), position)
        forms: dict[str, str] = {}
        with open(CACHE / "ngsl_teach.csv", encoding="utf-8") as f:
            for line in f:
                if line.startswith("#") or not line.strip():
                    continue
                words = [w.strip().lower() for w in line.strip().split(",") if w.strip()]
                lemma = words[0]
                for form in words:
                    best = forms.get(form)
                    if best is None or ranks.get(lemma, 10**6) < ranks.get(best, 10**6):
                        forms[form] = lemma
        return cls(ranks, forms)

    def lemma_of(self, token: str) -> str | None:
        """Lema da palavra da frase, ou None se nao e do vocabulario. Trata contracoes (don't, I'm, Tom's)."""
        if token in NAMES:
            return token
        base = token
        if "'" in token:
            base, _, suffix = token.partition("'")
            if suffix not in CONTRACTION_SUFFIXES and token not in self.forms:
                return None
        if base in NAMES:
            return base
        return self.forms.get(token) or self.forms.get(base)

    def required_rank(self, words: list[str]) -> int | None:
        """Posicao da palavra mais rara da frase (None se alguma nao e conhecida ou passa do vocabulario)."""
        worst = 0
        for word in words:
            lemma = self.lemma_of(word)
            if lemma is None:
                return None
            if lemma in NAMES:
                continue
            rank = self.rank.get(lemma)
            if rank is None or rank > MAX_RANK:
                return None
            worst = max(worst, rank)
        return worst or None


def lesson_index(rank: int) -> int:
    return (rank - 1) // LESSON_WORDS


def cefr_of(rank: int) -> str:
    if rank <= 600:
        return "A1"
    return "A2" if rank <= 1500 else "B1"
