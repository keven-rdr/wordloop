#!/usr/bin/env python3
"""Baixa as fontes abertas do lote v1 para content/.cache/ (fora do Git) e registra url, data e sha256.

Uso:  python content/tools/fetch_sources.py        (a partir da raiz do repositorio)

Fontes (ver docs/adr/0031-fontes-do-conteudo-v1.md):
  * NGSL 1.2 (lista de frequencia, CC BY-SA 4.0): lemas, ranking e formas flexionadas.
  * Tatoeba (frases e traducoes, CC BY 2.0 FR): ingles, portugues e os links entre eles.
O manifesto content/seed/v1/sources.json e versionado; os arquivos baixados, nao.
"""
import hashlib
import json
import sys
import urllib.request
from datetime import date

from common import CACHE, V1

NGSL = "https://www.newgeneralservicelist.com/s/"
TATOEBA = "https://downloads.tatoeba.org/exports/per_language/"
FILES = {
    "ngsl_stats.csv": NGSL + "NGSL_12_stats.csv",
    "ngsl_teach.csv": NGSL + "NGSL_12_lemmatized_for_teaching.csv",
    "eng_sentences.tsv.bz2": TATOEBA + "eng/eng_sentences_detailed.tsv.bz2",
    "por_sentences.tsv.bz2": TATOEBA + "por/por_sentences_detailed.tsv.bz2",
    "eng-por_links.tsv.bz2": TATOEBA + "eng/eng-por_links.tsv.bz2",
}


def download(name: str, url: str) -> dict:
    target = CACHE / name
    request = urllib.request.Request(url, headers={"User-Agent": "wordloop-content-tools"})
    with urllib.request.urlopen(request, timeout=300) as response:  # noqa: S310 (URL fixa, https)
        data = response.read()
    target.write_bytes(data)
    return {"file": name, "url": url, "sha256": hashlib.sha256(data).hexdigest(), "bytes": len(data)}


def main() -> int:
    CACHE.mkdir(parents=True, exist_ok=True)
    V1.mkdir(parents=True, exist_ok=True)
    entries = []
    for name, url in FILES.items():
        print(f"baixando {name} ...", flush=True)
        entries.append(download(name, url))
    manifest = {"retrieved_at": date.today().isoformat(), "files": entries}
    (V1 / "sources.json").write_text(json.dumps(manifest, indent=2) + "\n", encoding="utf-8", newline="\n")
    print(f"ok: {len(entries)} arquivos em {CACHE}; manifesto em {V1 / 'sources.json'}")
    return 0


if __name__ == "__main__":
    sys.exit(main())
