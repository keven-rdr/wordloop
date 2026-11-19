#!/usr/bin/env python3
"""Valida os dados do lote v1 (content/seed/v1/*.json). Falhou = nao se gera migracao nem se publica.

Uso:  python content/tools/validate.py        (a partir da raiz do repositorio; roda no CI e no `npm run check`)
"""
import json
import sys
import unicodedata

from common import SENTENCE_CHARS, V1, norm
from tatoeba import BLOCKED, PT_PT

POS = {"noun", "verb", "adj", "adv", "prep", "conj", "det", "pron"}
CEFR = {"A1", "A2", "B1", "B2", "C1", "C2"}
MIN_WORDS, MIN_SENTENCES = 2000, 2000
MIN_TILES, MAX_TILES = 3, 9  # pecas de "montar frase"


def _text_problems(label: str, value) -> list[str]:
    if not isinstance(value, str) or not value.strip():
        return [f"{label}: texto vazio"]
    out = []
    if value != value.strip() or "\n" in value or "\t" in value:
        out.append(f"{label}: espaco sobrando ou quebra de linha")
    if unicodedata.normalize("NFC", value) != value:
        out.append(f"{label}: texto fora da normalizacao NFC")
    return out


def _check_units(units: list[dict]) -> tuple[list[str], set[int]]:
    problems, positions = [], set()
    for unit in units:
        if unit["position"] in positions:
            problems.append(f"unidade repetida: {unit['position']}")
        positions.add(unit["position"])
        if not unit["theme_key"].startswith("lesson-"):
            problems.append(f"theme_key inesperada: {unit['theme_key']}")
    keys = [u["theme_key"] for u in units]
    if len(set(keys)) != len(keys):
        problems.append("theme_key repetida")
    return problems, positions


def _check_words(words: list[dict], positions: set[int]) -> list[str]:
    problems, seen = [], set()
    if len(words) < MIN_WORDS:
        problems.append(f"palavras: {len(words)} < {MIN_WORDS}")
    for w in words:
        label = f"palavra {w.get('lemma')!r}"
        if w["lemma"] in seen:
            problems.append(f"{label}: repetida")
        seen.add(w["lemma"])
        if w["pos"] not in POS:
            problems.append(f"{label}: pos invalido {w['pos']!r}")
        if w["cefr"] not in CEFR or w["unit"] not in positions:
            problems.append(f"{label}: cefr ou unidade invalidos")
        problems += _text_problems(f"{label} glosa", w["gloss"])
        if len(w["gloss"]) > 80:
            problems.append(f"{label}: glosa longa demais")
        ex = w.get("ex")
        if ex:
            problems += _text_problems(f"{label} exemplo", ex["text"]) + _text_problems(f"{label} exemplo pt", ex["pt"])
            if not (ex["eng_id"] and ex["por_id"]):
                problems.append(f"{label}: exemplo sem origem (ids do Tatoeba)")
    return problems


def sentence_problems(s: dict, positions: set[int]) -> list[str]:
    label = f"frase {s.get('text')!r}"
    text = s["text"]
    problems = _text_problems(label, text) + _text_problems(f"{label} pt", s["pt"])
    tiles = text.split()
    if not MIN_TILES <= len(tiles) <= MAX_TILES:
        problems.append(f"{label}: {len(tiles)} pecas (de {MIN_TILES} a {MAX_TILES})")
    if not SENTENCE_CHARS.fullmatch(text) or text[-1] not in ".?!" or any(m in text[:-1] for m in ".?!"):
        problems.append(f"{label}: caracteres ou pontuacao fora do padrao")
    if {w.strip(".,?!").lower() for w in tiles} & BLOCKED:
        problems.append(f"{label}: termo bloqueado")
    if any(mark in f" {s['pt'].lower()} " for mark in PT_PT):
        problems.append(f"{label}: traducao com marca de portugues de Portugal")
    if s["cefr"] not in CEFR or s["unit"] not in positions:
        problems.append(f"{label}: cefr ou unidade invalidos")
    if not (s["eng_id"] and s["por_id"]):
        problems.append(f"{label}: sem origem (ids do Tatoeba)")
    return problems


def _check_sentences(sentences: list[dict], positions: set[int]) -> list[str]:
    problems, seen = [], set()
    if len(sentences) < MIN_SENTENCES:
        problems.append(f"frases: {len(sentences)} < {MIN_SENTENCES}")
    for s in sentences:
        key = norm(s["text"])
        if key in seen:
            problems.append(f"frase repetida: {s['text']!r}")
        seen.add(key)
        problems += sentence_problems(s, positions)
    return problems


def validate(units: list[dict], words: list[dict], sentences: list[dict]) -> list[str]:
    unit_problems, positions = _check_units(units)
    return unit_problems + _check_words(words, positions) + _check_sentences(sentences, positions)


def load() -> tuple[list[dict], list[dict], list[dict]]:
    return tuple(json.loads((V1 / f"{n}.json").read_text(encoding="utf-8")) for n in ("units", "words", "sentences"))  # type: ignore[return-value]


def main() -> int:
    units, words, sentences = load()
    problems = validate(units, words, sentences)
    for p in problems[:50]:
        print("ERRO", p)
    if problems:
        print(f"validate: {len(problems)} problema(s)")
        return 1
    print(f"validate: ok ({len(units)} licoes, {len(words)} palavras, {len(sentences)} frases)")
    return 0


if __name__ == "__main__":
    sys.exit(main())
