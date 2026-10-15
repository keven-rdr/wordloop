import { describe, expect, it } from 'vitest';
import { footerParts } from './formatVersion';

const FULL = 'abc1234def5678abc1234def5678abc1234def56';

describe('footerParts', () => {
  it('build de release mostra versoes semver e um unico commit curto quando web e API sao o mesmo commit', () => {
    const parts = footerParts({ version: '0.1.0', commit: 'abc1234' }, { version: '0.1.0', commit: FULL, env: 'tst' });
    expect(parts).toEqual({ kind: 'version', web: 'v0.1.0', api: 'v0.1.0', env: 'tst', commit: 'abc1234' });
  });

  it('mostra os dois commits so quando de fato diferem', () => {
    const parts = footerParts(
      { version: '0.1.0', commit: 'abc1234' },
      { version: '0.1.0', commit: 'fffffff0000', env: 'tst' },
    );
    expect(parts.commit).toBe('abc1234/fffffff');
  });

  it('build comum (a versao e so o hash) mostra o commit curto no lugar da versao, sem repetir o hash', () => {
    const parts = footerParts(
      { version: 'abc1234', commit: 'abc1234' },
      { version: 'abc1234', commit: FULL, env: 'tst' },
    );
    expect(parts).toEqual({ kind: 'build', web: 'abc1234', api: 'abc1234', env: 'tst' });
  });

  it('um lado em release e o outro nao: usa o formato de release, com o commit curto no lugar da versao do outro', () => {
    const parts = footerParts(
      { version: 'abc1234', commit: 'abc1234' },
      { version: '0.2.0', commit: FULL, env: 'tst' },
    );
    expect(parts).toMatchObject({ kind: 'version', web: 'abc1234', api: 'v0.2.0' });
  });

  it('sem resposta da API mostra so o web', () => {
    expect(footerParts({ version: '0.1.0', commit: FULL })).toEqual({
      kind: 'versionWebOnly',
      web: 'v0.1.0',
      commit: 'abc1234',
    });
    expect(footerParts({ version: 'abc1234', commit: 'abc1234' })).toEqual({ kind: 'buildWebOnly', web: 'abc1234' });
  });

  it('versao como 0.1.0-5-gabc1234 (git describe) tambem e release', () => {
    expect(footerParts({ version: '0.1.0-5-gabc1234', commit: 'abc1234' }).web).toBe('v0.1.0-5-gabc1234');
  });
});
