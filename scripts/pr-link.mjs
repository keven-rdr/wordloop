#!/usr/bin/env node
// Monta o link da PR ja preenchida: abre o formulario do GitHub com o titulo e a descricao no lugar, sem CLI nem token.
// Uso: node scripts/pr-link.mjs --title "feat(api): ..." --body <arquivo.md> [--base develop] [--head <branch>]
// Sem --body, le .git/wl-pr/<branch>.md (fora do Git: a descricao nunca entra em commit).
import { execFileSync } from "node:child_process";
import { existsSync, readFileSync } from "node:fs";
import { join } from "node:path";

// O mesmo padrao do check obrigatorio "pr-title" (.github/workflows/pr-title.yml): o titulo vira o commit do squash.
const TITLE = /^(feat|fix|docs|style|refactor|perf|test|build|ci|chore|revert)(\([a-z0-9._/-]+\))?!?: .{3,100}$/;
// Acima disso o navegador ou o GitHub cortam a URL; a descricao entao vai para a area de transferencia a mao.
const MAX_URL = 8000;

const git = (...args) => execFileSync("git", args, { encoding: "utf8" }).trim();

function arg(name) {
  const at = process.argv.indexOf(`--${name}`);
  return at > 0 ? process.argv[at + 1] : undefined;
}

function fail(message) {
  console.error(`pr-link: ${message}`);
  process.exit(1);
}

// owner/repo a partir do remoto (https ou ssh).
function repoOf(remote) {
  const m = /github\.com[:/]([^/]+)\/(.+?)(?:\.git)?$/.exec(remote);
  if (!m) fail(`o remoto origin nao e do GitHub: ${remote}`);
  return `${m[1]}/${m[2]}`;
}

function pushed(head) {
  try {
    return git("rev-parse", `origin/${head}`) === git("rev-parse", "HEAD");
  } catch {
    return false;
  }
}

const head = arg("head") ?? git("rev-parse", "--abbrev-ref", "HEAD");
const base = arg("base") ?? "develop";
const title = arg("title");
const bodyFile = arg("body") ?? join(git("rev-parse", "--git-dir"), "wl-pr", `${head.replaceAll("/", "-")}.md`);

if (!title) fail("falta --title");
if (!TITLE.test(title)) fail(`titulo fora do Conventional Commits (o check pr-title reprova): "${title}"`);
if (head === base || ["develop", "main"].includes(head)) fail(`a PR sai de uma branch de trabalho, nao de ${head}`);
if (!existsSync(bodyFile)) fail(`descricao nao encontrada: ${bodyFile}`);
const body = readFileSync(bodyFile, "utf8").trim();
if (body.includes("<!--")) fail("a descricao ainda tem comentarios do template (<!-- ... -->): preencha cada secao");

const repo = repoOf(git("remote", "get-url", "origin"));
// A barra da branch fica como esta (feat/x): o GitHub espera o caminho, nao %2F.
const ref = (name) => name.split("/").map(encodeURIComponent).join("/");
const query = new URLSearchParams({ expand: "1", title, body }).toString().replaceAll("+", "%20");
const url = `https://github.com/${repo}/compare/${ref(base)}...${ref(head)}?${query}`;

if (!pushed(head)) console.warn(`aviso: origin/${head} nao esta no mesmo commit que o HEAD; faca o push antes de abrir.`);
if (url.length > MAX_URL) {
  const short = `https://github.com/${repo}/compare/${ref(base)}...${ref(head)}?${new URLSearchParams({ expand: "1", title }).toString().replaceAll("+", "%20")}`;
  console.warn(`aviso: com a descricao o link passa de ${MAX_URL} caracteres; abre so com o titulo. Cole a descricao de ${bodyFile}.`);
  console.log(short);
} else {
  console.log(url);
}
