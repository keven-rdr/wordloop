import { describe, expect, it } from 'vitest';
import { detectLocale } from './i18n';

describe('detectLocale', () => {
  it.each([
    [['pt-BR', 'pt', 'en-US', 'en'], 'pt-BR'],
    [['pt-BR', 'en'], 'pt-BR'],
    [['pt-PT'], 'pt-BR'],
    [['en-GB', 'pt-BR'], 'en-US'],
    [['en-US'], 'en-US'],
    [['es', 'pt-BR'], 'pt-BR'],
    [['es', 'fr'], 'pt-BR'],
    [[], 'pt-BR'],
  ])('%j -> %s', (languages, expected) => {
    expect(detectLocale(languages)).toBe(expected);
  });
});
