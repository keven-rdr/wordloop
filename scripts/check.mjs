#!/usr/bin/env node
// `npm run check`: Definition of Done local. Roda nos componentes alterados contra a base (develop);
// `--all` roda nos dois. Mesmos comandos do CI (api-ci, web-ci), sem Sonar/E2E/Lighthouse.
import { execFileSync, spawnSync } from "node:child_process";

const all = process.argv.includes("--all");
const git = (...a) => execFileSync("git", a, { encoding: "utf8" }).trim();

function changedComponents() {
  const base = ["origin/develop", "develop", "origin/main", "main"].find((r) => {
    try { git("rev-parse", "--verify", r); return true; } catch { return false; }
  });
  if (!base) return new Set(["api", "web"]);
  const mergeBase = git("merge-base", "HEAD", base);
  const files = [
    ...git("diff", "--name-only", mergeBase).split("\n"),
    ...git("ls-files", "--others", "--exclude-standard").split("\n"),
  ];
  const set = new Set();
  for (const f of files) {
    if (f.startsWith("api/")) set.add("api");
    if (f.startsWith("web/")) set.add("web");
    if (f.startsWith("api/openapi/")) set.add("web"); // o cliente web e gerado do contrato
  }
  return set;
}

const steps = {
  api: [
    ["api", "gofmt", ["-l", "."], { failIfOutput: true }],
    ["api", "go", ["vet", "./..."]],
    ["api", "golangci-lint", ["run"]],
    ["api", "go", ["run", "./scripts/check-structure"]],
    ["api", "go-arch-lint", ["check"]],
    ["api", "go", ["test", "-short", "./..."]],
  ],
  web: [
    ["web", "npx", ["biome", "ci", "./src", "./scripts"]],
    ["web", "npx", ["eslint", "./src"]],
    ["web", "npx", ["tsc", "-b"]],
    ["web", "npm", ["run", "check:i18n"]],
    ["web", "npm", ["run", "check:structure"]],
    ["web", "npx", ["vitest", "run"]],
  ],
};

const targets = all ? new Set(["api", "web"]) : changedComponents();
if (targets.size === 0) console.log("check: nenhum componente alterado (use --all para tudo)");

let failed = 0;
for (const t of ["api", "web"].filter((c) => targets.has(c))) {
  for (const [cwd, cmd, args, opts] of steps[t]) {
    const label = `${cwd}: ${cmd} ${args.join(" ")}`;
    const r = spawnSync(cmd, args, { cwd, encoding: "utf8", shell: process.platform === "win32" });
    const bad = r.status !== 0 || (opts?.failIfOutput && r.stdout.trim() !== "");
    console.log(`${bad ? "FALHOU" : "ok    "}  ${label}`);
    if (bad) {
      failed++;
      process.stdout.write(`${r.stdout ?? ""}${r.stderr ?? ""}`);
    }
  }
}
if (targets.has("web") || targets.has("api")) {
  const r = spawnSync("node", ["scripts/check-changed-lines.mjs"], { encoding: "utf8" });
  console.log(`${r.status === 0 ? "ok    " : "FALHOU"}  check-changed-lines`);
  if (r.status !== 0) { failed++; process.stdout.write(`${r.stdout}${r.stderr}`); }
}
console.log(failed === 0 ? "\ncheck: tudo verde" : `\ncheck: ${failed} etapa(s) falharam`);
process.exit(failed === 0 ? 0 : 1);
