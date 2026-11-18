export type Section = Readonly<{ title: string; body: readonly string[] }>;

const isSection = (value: unknown): value is Section =>
  typeof value === 'object' &&
  value !== null &&
  'title' in value &&
  typeof value.title === 'string' &&
  'body' in value &&
  Array.isArray(value.body) &&
  value.body.every((line) => typeof line === 'string');

// As seções vêm do arquivo de idioma como uma lista de { title, body[] }; o que não tiver essa forma é ignorado.
export const sectionsOf = (value: unknown): Section[] => (Array.isArray(value) ? value.filter(isSection) : []);
