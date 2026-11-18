import { describe, expect, it } from 'vitest';
import { sectionsOf } from './legal';

describe('sectionsOf', () => {
  it('mantém as seções bem formadas', () => {
    const sections = [
      { title: 'A', body: ['um', 'dois'] },
      { title: 'B', body: [] },
    ];
    expect(sectionsOf(sections)).toEqual(sections);
  });

  it('ignora o que não tem a forma de uma seção', () => {
    expect(
      sectionsOf([
        { title: 'ok', body: ['x'] },
        { title: 'sem corpo' },
        { body: ['sem título'] },
        { title: 'corpo errado', body: [1, 2] },
        { title: 'corpo texto', body: 'x' },
        'texto',
        null,
        42,
      ]),
    ).toEqual([{ title: 'ok', body: ['x'] }]);
  });

  it('o que não é lista vira lista vazia (chave que faltou no idioma)', () => {
    expect(sectionsOf('legal.privacy.sections')).toEqual([]);
    expect(sectionsOf(undefined)).toEqual([]);
    expect(sectionsOf({ title: 'A', body: [] })).toEqual([]);
  });
});
