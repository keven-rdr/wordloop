"""Apoio comum das ferramentas de conteudo (Python, so biblioteca padrao)."""
import pathlib
import re
import unicodedata

ROOT = pathlib.Path(__file__).resolve().parents[2]
CACHE = ROOT / "content" / ".cache"  # fontes baixadas: nunca vao para o Git
V1 = ROOT / "content" / "seed" / "v1"  # dados do lote v1 (fonte da verdade, versionados)
MIGRATIONS = ROOT / "api" / "migrations"

# Nomes proprios liberados nas frases (o Tatoeba e cheio deles); contam como palavra conhecida.
NAMES = {"tom", "mary"}

TOKEN = re.compile(r"[a-z]+(?:'[a-z]+)?")
SENTENCE_CHARS = re.compile(r"[A-Za-z' ,.?!]+")


def norm(text: str) -> str:
    """Chave de comparacao: minusculas, sem pontuacao, espacos unicos."""
    return " ".join(TOKEN.findall(unicodedata.normalize("NFKC", text).lower()))


def tokens(text: str) -> list[str]:
    return TOKEN.findall(text.lower())


def q(value) -> str:
    """Literal SQL: NULL, numero ou texto entre aspas simples."""
    if value is None:
        return "NULL"
    if isinstance(value, (int, float)):
        return str(value)
    return "'" + str(value).replace("'", "''") + "'"
