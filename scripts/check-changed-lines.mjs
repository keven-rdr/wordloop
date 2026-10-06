#!/usr/bin/env node
// ESQUELETO (Fase 4): aplica regras nas LINHAS ALTERADAS contra a base (develop; se não existir, main).
// Ideia (a mesma do Contratos, escrita do zero): git diff -U0 -> mapa arquivo -> linhas novas -> regras por regex.
// Meta: o mesmo critério do Sonar, antes do CI e sem acusar dívida antiga. Uso: node scripts/check-changed-lines.mjs [--base=develop]
import { execFileSync } from "node:child_process";
import { readFileSync } from "node:fs";

const git = (...a) => execFileSync("git", a, { encoding: "utf8" }).trim();
const baseArg = process.argv.find((a) => a.startsWith("--base="))?.slice(7);
const base = [baseArg, "origin/develop", "develop", "origin/main", "main"].filter(Boolean).find((r) => {
  try { git("rev-parse", "--verify", r); return true; } catch { return false; }
});
if (!base) { console.error("base não encontrada (develop/main)"); process.exit(2); }
const mergeBase = git("merge-base", "HEAD", base);

// arquivo -> Set(números das linhas adicionadas/alteradas)
const changed = new Map();
const diff = git("diff", "-U0", "--diff-filter=AM", mergeBase, "--", "*.ts", "*.tsx", "*.go");
let file = null;
for (const line of diff.split("\n")) {
  if (line.startsWith("+++ b/")) { file = line.slice(6); changed.set(file, new Set()); continue; }
  const m = /^@@ -\d+(?:,\d+)? \+(\d+)(?:,(\d+))? @@/.exec(line);
  if (m && file) {
    const start = Number(m[1]); const n = m[2] === undefined ? 1 : Number(m[2]);
    for (let i = 0; i < n; i++) changed.get(file).add(start + i);
  }
}

// Regras (exemplos; completar com a lista do web/AGENTS.md). Cada regra: id, aplica-se a, teste por linha, mensagem.
const rules = [
  { id: "S6606", ext: /\.tsx?$/, test: (l) => /\|\|\s*(""|''|0|\[\]|\{\})/.test(l), msg: "use ?? em vez de || para valor nulável" },
  { id: "S3358", ext: /\.tsx?$/, test: (l) => /\?[^?:]*\?[^:]*:/.test(l) && /:/.test(l), msg: "ternário aninhado" },
  { id: "S4325", ext: /\.tsx?$/, test: (l) => /\sas\s+(string|number|boolean)\b/.test(l), msg: "`as` possivelmente desnecessário" },
];

let violations = 0;
for (const [path, lines] of changed) {
  if (/(^|\/)(gen|generated)\//.test(path) || /\.gen\./.test(path)) continue;
  const src = readFileSync(path, "utf8").split("\n");
  for (const n of lines) {
    for (const r of rules) {
      if (r.ext.test(path) && r.test(src[n - 1] ?? "")) {
        console.error(`${path}:${n}  [${r.id}] ${r.msg}`);
        violations++;
      }
    }
  }
}
console.log(violations === 0 ? "check-changed-lines: ok" : `check-changed-lines: ${violations} violação(ões) em linhas alteradas`);
process.exit(violations === 0 ? 0 : 1);
