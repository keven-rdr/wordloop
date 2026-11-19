#!/usr/bin/env python3
"""Monta os dados do lote v1 (content/seed/v1/*.json) a partir das fontes em content/.cache e das glosas escritas a mao.

Uso:  python content/tools/build_v1.py        (depois de fetch_sources.py; a partir da raiz do repositorio)
Saida: content/seed/v1/units.json, words.json e sentences.json (fonte da verdade, versionada e conferivel em diff).

Regras (ver docs/features/2026-11-18-conteudo-e-montar-frases.md):
  * Vocabulario: os 2.100 lemas mais frequentes da NGSL, menos os que ja estao no v0.
  * Licao: 30 posicoes do ranking de frequencia; a licao n vira a unidade de posicao 6 + n.
  * Frase: entra na licao da palavra mais rara que usa; ate 30 por licao.
  * Exemplo de cada palavra: frase curta do Tatoeba so com palavras tao frequentes quanto ela.
"""
import json
import sys
from pathlib import Path

from common import V1, norm
from ngsl import MAX_RANK, Ngsl, cefr_of, lesson_index
from pick import candidates, pick_examples, pick_sentences
from tatoeba import Pair, load_pairs

SEED_V0 = Path(__file__).resolve().parents[1] / "seed" / "seed-v0.json"
FIRST_POSITION = 7  # as unidades 1 a 6 sao do v0


def read_glosses() -> dict[str, tuple[str, str]]:
    out: dict[str, tuple[str, str]] = {}
    for path in sorted((V1 / "glosses").glob("*.tsv")):
        for number, line in enumerate(path.read_text(encoding="utf-8").splitlines(), start=1):
            if not line.strip():
                continue
            parts = line.split("|")
            if len(parts) != 3 or not all(p.strip() for p in parts):
                raise SystemExit(f"{path.name}:{number}: esperava lema|pos|glosa, veio {line!r}")
            lemma, pos, gloss = (p.strip() for p in parts)
            if lemma in out:
                raise SystemExit(f"{path.name}:{number}: lema repetido: {lemma}")
            out[lemma] = (pos, gloss)
    return out


def source_of(pair: Pair) -> dict:
    return {"eng_id": pair.eng_id, "por_id": pair.por_id, "eng_user": pair.eng_user, "por_user": pair.por_user}


def build_units(lessons: list[int]) -> tuple[list[dict], dict[int, int]]:
    """Unidades `lesson-<nivel>-<nn>`: numeradas dentro de cada nivel CEFR; posicao segue a ordem de frequencia."""
    counters: dict[str, int] = {}
    units, position_of = [], {}
    for index in sorted(set(lessons)):
        level = cefr_of(index * 30 + 1).lower()
        counters[level] = counters.get(level, 0) + 1
        position = FIRST_POSITION + len(units)
        units.append({"position": position, "theme_key": f"lesson-{level}-{counters[level]:02d}"})
        position_of[index] = position
    return units, position_of


def main() -> int:
    ngsl = Ngsl.load()
    glosses = read_glosses()
    seed_v0 = json.loads(SEED_V0.read_text(encoding="utf-8"))
    v0 = {w["lemma"].lower() for w in seed_v0["words"]}
    v0_sentences = {norm(s["text"]) for s in seed_v0["sentences"]}
    ranked = [(rank, lemma) for lemma, rank in sorted(ngsl.rank.items(), key=lambda kv: kv[1]) if rank <= MAX_RANK]
    new_words = [(rank, lemma) for rank, lemma in ranked if lemma not in v0]
    missing = [lemma for _, lemma in new_words if lemma not in glosses]
    extra = [lemma for lemma in glosses if lemma not in {l for _, l in new_words}]
    if missing or extra:
        raise SystemExit(f"glosas: faltam {missing[:10]} (total {len(missing)}), sobram {extra[:10]} (total {len(extra)})")

    cands = [c for c in candidates(load_pairs(), ngsl) if norm(c.pair.eng) not in v0_sentences]
    sentences = pick_sentences(cands, ngsl)
    examples = pick_examples(ranked, cands, {c.pair.eng_id for c in sentences})
    units, position_of = build_units([lesson_index(r) for r, _ in new_words] + [c.lesson for c in sentences])

    words = []
    for rank, lemma in new_words:
        pos, gloss = glosses[lemma]
        entry = {"rank": rank, "lemma": lemma, "pos": pos, "gloss": gloss, "cefr": cefr_of(rank), "unit": position_of[lesson_index(rank)]}
        example = examples.get(lemma)
        if example:
            entry["ex"] = {"text": example.pair.eng, "pt": example.pair.por, **source_of(example.pair)}
        words.append(entry)
    out_sentences = [
        {"text": c.pair.eng, "pt": c.pair.por, "cefr": cefr_of(c.rank), "unit": position_of[c.lesson], "rank": c.rank, **source_of(c.pair)}
        for c in sorted(sentences, key=lambda c: (c.rank, c.pair.eng_id))
    ]
    for name, data in (("units", units), ("words", words), ("sentences", out_sentences)):
        (V1 / f"{name}.json").write_text(json.dumps(data, ensure_ascii=False, indent=1) + "\n", encoding="utf-8")
    with_example = sum(1 for w in words if "ex" in w)
    print(f"ok: {len(units)} licoes, {len(words)} palavras ({with_example} com exemplo), {len(out_sentences)} frases")
    return 0


if __name__ == "__main__":
    sys.exit(main())
