#!/usr/bin/env node
// Guardrail PreToolUse: bloqueia (exit 2) o que NÃO pode acontecer. Regras escritas em AGENTS.md.
// Entrada: JSON no stdin ({ tool_name, tool_input, cwd }). Saída de bloqueio: mensagem em stderr + exit 2.
import { existsSync, readFileSync } from "node:fs";
import { relative, sep } from "node:path";

const MAX_LINES = 300;
const input = JSON.parse(readFileSync(0, "utf8"));
const tool = input.tool_name;
const ti = input.tool_input ?? {};
const root = process.env.CLAUDE_PROJECT_DIR ?? input.cwd ?? process.cwd();
const file = ti.file_path ? relative(root, ti.file_path).split(sep).join("/") : "";

const deny = (msg) => {
  console.error(`BLOQUEADO: ${msg}`);
  process.exit(2);
};

// 1) Segredos: nunca ler nem editar .env* (exceto *.example)
if (/(^|\/)\.env(\.|$)/.test(file) && !file.endsWith(".example")) {
  deny("arquivos .env* não são lidos nem editados; use o *.example.");
}
if (tool === "Read") process.exit(0);

// 2) Código gerado: altere a fonte (OpenAPI/SQL) e rode `npm run gen`
if (/(^|\/)(gen|generated)\//.test(file) || /routeTree\.gen\.ts$/.test(file) || /\/db\/[^/]*\.sql\.go$/.test(file)) {
  deny("código gerado: altere api/openapi/openapi.yaml ou o .sql e rode `npm run gen`.");
}

// 3) Texto final do arquivo após a edição
let after = ti.content;
if (tool === "Edit" && ti.file_path && existsSync(ti.file_path)) {
  const cur = readFileSync(ti.file_path, "utf8");
  after = ti.replace_all
    ? cur.split(ti.old_string).join(ti.new_string)
    : cur.replace(ti.old_string, () => ti.new_string);
}
if (typeof after !== "string") process.exit(0);

const isCode = /\.(go|ts|tsx)$/.test(file) && !/\.gen\./.test(file);
const isTest = /(_test\.go|\.test\.tsx?|\.spec\.tsx?|^web\/e2e\/)/.test(file);

// 4) Tamanho do arquivo
if (isCode && after.split("\n").length > MAX_LINES) {
  deny(`arquivo passaria de ${MAX_LINES} linhas (${file}). Divida por funcionalidade.`);
}

// 5) Web: string de UI crua e cor crua
if (/^web\/src\/.*\.tsx$/.test(file) && !isTest) {
  const jsxText = /<[A-Za-z][^>]*>\s*[^<>{}\s][^<>{}]*<\/[A-Za-z]/;
  const attr = /\b(title|placeholder|aria-label|alt)="[^"{}]+"/;
  for (const line of after.split("\n")) {
    if (jsxText.test(line) || attr.test(line)) {
      deny(`texto de interface literal em ${file}: use t("chave") e adicione a chave em pt-BR e en-US.\n  > ${line.trim()}`);
    }
  }
}
if (/^web\/src\/.*\.(ts|tsx)$/.test(file) && !/tokens\//.test(file) && !isTest && /#[0-9a-fA-F]{3,8}\b/.test(after)) {
  deny(`cor crua em ${file}: use tokens StyleX (shared/ui/tokens).`);
}

// 6) API: erro de API com string literal
if (/^api\/internal\/.*\.go$/.test(file) && !isTest && /\bproblem\.\w+\(\s*"/.test(after)) {
  deny(`erro de API com string literal em ${file}: use a constante gerada de chave (problem.New(CodeXxx, ...)).`);
}

process.exit(0);
