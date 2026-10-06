#!/usr/bin/env python3
"""Valida o seed da amostra num SQLite em memória (NÃO é PostgreSQL).

Prova: integridade referencial, CHECKs, idempotência (2 execuções = mesmas contagens) e as consultas de
conferência. NÃO prova a sintaxe do PostgreSQL: isso exige rodar o DDL e o seed num Postgres (Sprint 0).
"""
import pathlib
import re
import sqlite3
import sys

here = pathlib.Path(__file__).parent
ddl = (here.parent.parent / "fase-2" / "content-schema.sql").read_text(encoding="utf-8")

# Adaptação mínima de dialeto PostgreSQL -> SQLite
ddl = re.sub(r"DEFAULT gen_random_uuid\(\)", "", ddl)
ddl = re.sub(r"\buuid\b", "TEXT", ddl)
ddl = re.sub(r"\btimestamptz\b", "TEXT", ddl)
ddl = ddl.replace("DEFAULT now()", "DEFAULT CURRENT_TIMESTAMP")

seed = (here / "out" / "sample-seed.sql").read_text(encoding="utf-8").replace("now()", "CURRENT_TIMESTAMP")
verify = (here / "out" / "sample-verify.sql").read_text(encoding="utf-8")

con = sqlite3.connect(":memory:")
con.execute("PRAGMA foreign_keys = ON")
con.executescript(ddl)


def counts() -> dict:
    tabs = ["source", "subject", "course", "unit", "lexeme", "item", "item_text", "item_relation", "media", "item_media", "attribution"]
    return {t: con.execute(f"SELECT COUNT(*) FROM {t}").fetchone()[0] for t in tabs}


con.executescript(seed)
first = counts()
con.executescript(seed)          # 2ª execução: idempotência
second = counts()
assert first == second, f"seed NÃO é idempotente:\n{first}\n{second}"

failed = False
for stmt in [s for s in verify.split(";\n") if s.strip() and not s.strip().startswith("--") or "SELECT" in s]:
    sql = "\n".join(l for l in stmt.splitlines() if not l.strip().startswith("--")).strip().rstrip(";")
    if not sql.upper().startswith("SELECT"):
        continue
    rows = con.execute(sql).fetchall()
    if "AS violation" in sql:
        if rows:
            failed = True
            print("VIOLAÇÃO:", rows[:5])
    else:
        r = rows[0]
        flag = "ok " if r[3] == "ok" else "FALHA"
        failed = failed or r[3] != "ok"
        print(f"{flag}  {r[0]:<22} encontrado={r[1]:<4} esperado={r[2]}")

print("contagens após 2 execuções:", second)
print("RESULTADO:", "FALHOU" if failed else "tudo ok (SQLite; Postgres ainda não validado)")
sys.exit(1 if failed else 0)
