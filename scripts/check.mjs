#!/usr/bin/env node
// `npm run check`: Definition of Done local. Roda nos componentes alterados contra a base (develop);
// `--all` roda nos dois. Mesmos comandos do CI (api-ci, web-ci), sem Sonar/E2E/Lighthouse.
// Funciona numa maquina limpa: gera o codigo que falta e usa `go run` quando o binario nao esta no PATH.
import { execFileSync, spawnSync } from "node:child_process";
import { existsSync } from "node:fs";

// Versoes fixas das ferramentas de lint. MANTENHA IGUAL ao `version:` do golangci-lint-action em .github/workflows/api-ci.yml.
const GOLANGCI_LINT = "v2.14.0";
const GO_ARCH_LINT = "v1.19.0";

const all = process.argv.includes("--all");
const win = process.platform === "win32";
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

function onPath(cmd) {
  const r = spawnSync(win ? "where" : "which", [cmd], { encoding: "utf8", shell: win });
  return r.status === 0;
}

// Usa o binario do PATH; senao `go run <pacote>@<versao>` (1a execucao demora: compila; depois usa cache).
function goTool(bin, pkg, version, args) {
  return onPath(bin) ? [bin, args] : ["go", ["run", `${pkg}@${version}`, ...args]];
}

const [lintCmd, lintArgs] = goTool("golangci-lint", "github.com/golangci/golangci-lint/v2/cmd/golangci-lint", GOLANGCI_LINT, ["run"]);
const [archCmd, archArgs] = goTool("go-arch-lint", "github.com/fe3dback/go-arch-lint", GO_ARCH_LINT, ["check"]);

const prepare = {
  api: [{ missing: "api/internal/api/gen/api.gen.go", cwd: "api", cmd: "go", args: ["generate", "./..."], why: "codigo gerado da API" }],
  web: [
    { missing: "web/node_modules", cwd: "web", cmd: "npm", args: ["ci"], why: "dependencias do web" },
    { missing: "web/src/api/generated", cwd: "web", cmd: "npm", args: ["run", "gen:api"], why: "cliente gerado do contrato" },
  ],
};

const steps = {
  api: [
    ["api", "gofmt", ["-l", "."], { failIfOutput: true }],
    ["api", "go", ["vet", "./..."]],
    ["api", lintCmd, lintArgs],
    ["api", "go", ["run", "./scripts/check-structure"]],
    ["api", archCmd, archArgs],
    ["api", "go", ["test", "-short", "./..."]],
  ],
  web: [
    ["web", "npx", ["biome", "ci", "./src", "./scripts", "./e2e"]],
    ["web", "npx", ["eslint", "./src"]],
    ["web", "npx", ["tsc", "-b"]],
    ["web", "npm", ["run", "check:i18n"]],
    ["web", "npm", ["run", "check:structure"]],
    ["web", "npx", ["vitest", "run"]],
  ],
};

const run = (cwd, cmd, args) => spawnSync(cmd, args, { cwd, encoding: "utf8", shell: win });

const targets = all ? new Set(["api", "web"]) : changedComponents();
if (targets.size === 0) console.log("check: nenhum componente alterado (use --all para tudo)");

let failed = 0;
for (const t of ["api", "web"].filter((c) => targets.has(c))) {
  for (const p of prepare[t]) {
    if (existsSync(p.missing)) continue;
    console.log(`prep    ${p.why}: ${p.cmd} ${p.args.join(" ")}`);
    const r = run(p.cwd, p.cmd, p.args);
    if (r.status !== 0) { failed++; process.stdout.write(`${r.stdout ?? ""}${r.stderr ?? ""}`); }
  }
  for (const [cwd, cmd, args, opts] of steps[t]) {
    const label = `${cwd}: ${cmd} ${args.join(" ")}`;
    const r = run(cwd, cmd, args);
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
