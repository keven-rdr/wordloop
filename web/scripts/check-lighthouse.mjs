// Confere as notas do Lighthouse (JSON) contra os limites minimos e sai com erro se alguma ficar abaixo.
// Uso: node scripts/check-lighthouse.mjs lh-login.json [lh-signup.json ...]
import { readFileSync } from 'node:fs';

// Limites de 0 a 1. Desempenho com folga porque o runner do CI e mais lento e instavel que uma maquina de dev.
const MIN = { performance: 0.6, accessibility: 0.9, 'best-practices': 0.9, seo: 0.9 };

const files = process.argv.slice(2);
if (files.length === 0) {
  console.error('uso: node scripts/check-lighthouse.mjs <resultado.json> [...]');
  process.exit(2);
}

let failed = false;
for (const file of files) {
  const report = JSON.parse(readFileSync(file, 'utf8'));
  console.log(
    `\n${report.finalDisplayedUrl ?? report.requestedUrl} (Lighthouse ${report.lighthouseVersion}, ${report.configSettings.formFactor})`,
  );
  for (const [category, min] of Object.entries(MIN)) {
    const score = report.categories[category]?.score;
    const ok = typeof score === 'number' && score >= min;
    failed ||= !ok;
    const shown = typeof score === 'number' ? Math.round(score * 100) : 'sem nota';
    console.log(
      `  ${ok ? 'ok     ' : 'ABAIXO '} ${category.padEnd(15)} ${String(shown).padStart(4)}  (minimo ${Math.round(min * 100)})`,
    );
    if (!ok)
      console.log(`::error title=Lighthouse::${category} em ${shown}, minimo ${Math.round(min * 100)} (${file})`);
  }
}
process.exit(failed ? 1 : 0);
