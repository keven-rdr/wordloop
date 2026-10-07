// Impoe a estrutura do front: sem pastas por tipo tecnico na raiz de src/ (AGENTS.md, regra 6).
import { readdirSync, statSync } from 'node:fs';
import { join } from 'node:path';

const src = new URL('../src/', import.meta.url).pathname.replace(/^\/([A-Za-z]:)/, '$1');
const forbidden = new Set(['services', 'queries', 'schemas', 'utils', 'helpers', 'controllers']);

const problems = readdirSync(src)
  .filter((name) => statSync(join(src, name)).isDirectory() && forbidden.has(name))
  .map((name) => `src/${name}/: pasta por tipo tecnico e proibida; organize por funcionalidade`);

for (const p of problems) console.error(p);
console.log(problems.length === 0 ? 'check-structure: ok' : `check-structure: ${problems.length} problema(s)`);
process.exit(problems.length === 0 ? 0 : 1);
