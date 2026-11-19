"""Pares ingles <-> portugues do Tatoeba (CC BY 2.0 FR) ja filtrados por qualidade."""
import bz2
from dataclasses import dataclass

from common import CACHE, SENTENCE_CHARS, tokens

# Termos que nao entram num app de estudo aberto a todos (violencia, sexo, drogas, palavroes, politica).
BLOCKED = {
    "fuck", "fucking", "shit", "bitch", "damn", "ass", "asshole", "dick", "sex", "sexy", "naked", "nude", "porn",
    "rape", "kill", "killed", "murder", "suicide", "terrorist", "bomb", "drug", "drugs", "cocaine", "heroin",
    "gun", "weapon", "war", "gay", "jew", "muslim", "christian", "god", "jesus", "allah", "church", "nazi",
    "trump", "obama", "stupid", "idiot", "hate", "die", "dead", "death", "blood", "hell", "ugly", "fat", "drunk",
    "beer", "wine", "alcohol", "cigarette", "smoke", "cancer", "pregnant", "divorce", "prison", "police",
}
# Marcas de portugues de Portugal (o app ensina para falantes de pt-BR): descarta a traducao.
PT_PT = (
    " tu ", "estás", "estou a ", " a fazer", " a ler", " a comer", " a dormir", "autocarro", "comboio", "telemóvel",
    "casa de banho", "pequeno-almoço", "miúdo", "rapariga", "ecrã", "frigorífico", "equipa", "ficheiro", "rato ",
    "gajo", "pá,", "bué", "fixe", "giro", "chávena", "sumo", "contigo", "consigo ", "vou-me", "diz-me", "dá-me",
    "deixa-me", "faz-me", "-te ", "-se ", "-lhe", "-nos ", "-o.", "-a.", "-os.", "-as.",
)


@dataclass(frozen=True)
class Pair:
    eng_id: int
    eng: str
    eng_user: str
    por_id: int
    por: str
    por_user: str

    @property
    def words(self) -> list[str]:
        return tokens(self.eng)


def _rows(name: str):
    with bz2.open(CACHE / name, "rt", encoding="utf-8") as f:
        for line in f:
            yield line.rstrip("\n").split("\t")


def _eng_ok(text: str) -> bool:
    if not SENTENCE_CHARS.fullmatch(text) or text[-1] not in ".?!" or text.count(" ") < 2:
        return False
    if any(mark in text[:-1] for mark in ".?!"):  # uma frase so (nada de duas num item)
        return False
    words = tokens(text)
    return 3 <= len(words) <= 9 and not (set(words) & BLOCKED) and text[0].isupper()


def _por_ok(text: str, eng: str) -> bool:
    low = f" {text.lower()} "
    if any(mark in low for mark in PT_PT) or len(text) > 140 or text[-1] not in ".?!":
        return False
    ratio = len(text.split()) / max(1, len(eng.split()))
    return 0.5 <= ratio <= 2.6 and _same_names(text, eng)


def _same_names(por: str, eng: str) -> bool:
    """A traducao tem de falar da mesma pessoa: Tom <-> Tom, Mary <-> Maria/Mary (e nao inventar um dos dois)."""
    for en_name, pt_names in (("Tom", ("Tom",)), ("Mary", ("Mary", "Maria"))):
        in_eng = en_name in eng.split() or any(w.startswith(en_name + "'") for w in eng.split())
        in_por = any(n in por for n in pt_names)
        if in_eng != in_por:
            return False
    return True


def load_pairs() -> list[Pair]:
    links: dict[int, list[int]] = {}
    for row in _rows("eng-por_links.tsv.bz2"):
        if len(row) == 2:
            links.setdefault(int(row[0]), []).append(int(row[1]))
    wanted_por = {p for ps in links.values() for p in ps}
    eng = {int(r[0]): (r[2], r[3]) for r in _rows("eng_sentences.tsv.bz2") if len(r) >= 4 and int(r[0]) in links and _eng_ok(r[2])}
    por = {int(r[0]): (r[2], r[3]) for r in _rows("por_sentences.tsv.bz2") if len(r) >= 4 and int(r[0]) in wanted_por}
    pairs, seen = [], set()
    for eng_id in sorted(eng):
        text, user = eng[eng_id]
        key = " ".join(tokens(text))
        if key in seen:
            continue
        for por_id in links[eng_id]:
            if por_id in por and _por_ok(por[por_id][0], text):
                pairs.append(Pair(eng_id, text, user, por_id, por[por_id][0], por[por_id][1]))
                seen.add(key)
                break
    return pairs
