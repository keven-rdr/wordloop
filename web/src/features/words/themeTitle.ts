import type { TFunction } from 'i18next';

// As lições do lote v1 chegam como `lesson-a1-03`: uma chave só no i18n (nível e número como parâmetros)
// em vez de uma tradução por lição. Os temas antigos (`home-and-family`...) continuam com a chave própria.
const LESSON = /^lesson-([a-z]\d)-(\d+)$/;

export function themeTitle(t: TFunction<'words'>, themeKey: string): string {
  const [, level, n] = LESSON.exec(themeKey) ?? [];
  if (level && n) return t('units.lesson', { level: level.toUpperCase(), n: Number(n) });
  return t(`units.theme.${themeKey}`, { defaultValue: themeKey });
}
