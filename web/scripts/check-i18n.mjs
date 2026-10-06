// Garante que cada namespace existe em todos os idiomas e com as mesmas chaves.
import { readdirSync, readFileSync } from 'node:fs';
import { join } from 'node:path';

const root = new URL('../src/locales/', import.meta.url).pathname.replace(/^\/([A-Za-z]:)/, '$1');
const langs = readdirSync(root);

function keys(obj, prefix = '') {
  return Object.entries(obj).flatMap(([k, v]) =>
    typeof v === 'object' && v !== null ? keys(v, `${prefix}${k}.`) : [`${prefix}${k}`],
  );
}

const byLang = new Map(
  langs.map((lang) => [
    lang,
    new Map(
      readdirSync(join(root, lang))
        .filter((f) => f.endsWith('.json'))
        .map((f) => [f, new Set(keys(JSON.parse(readFileSync(join(root, lang, f), 'utf8'))))]),
    ),
  ]),
);

const namespaces = new Set([...byLang.values()].flatMap((m) => [...m.keys()]));
const problems = [];
for (const ns of namespaces) {
  const reference = new Set([...byLang.values()].flatMap((m) => [...(m.get(ns) ?? [])]));
  for (const [lang, files] of byLang) {
    const have = files.get(ns);
    if (!have) {
      problems.push(`${lang}/${ns}: arquivo ausente`);
      continue;
    }
    for (const key of reference) if (!have.has(key)) problems.push(`${lang}/${ns}: falta a chave "${key}"`);
  }
}

for (const p of problems) console.error(p);
console.log(problems.length === 0 ? 'check-i18n: ok' : `check-i18n: ${problems.length} problema(s)`);
process.exit(problems.length === 0 ? 0 : 1);
