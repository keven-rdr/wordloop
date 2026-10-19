import { describe, expect, it } from 'vitest';
import { badgeParts } from './formatVersion';

const FULL = 'abc1234def5678abc1234def5678abc1234def56';

describe('badgeParts', () => {
  it('release mostra versoes semver e um unico commit curto quando web e API sao o mesmo commit', () => {
    const parts = badgeParts({ version: '0.1.0', commit: 'abc1234' }, { version: '0.1.0', commit: FULL, env: 'tst' });
    expect(parts).toEqual({ kind: 'version', web: 'v0.1.0', api: 'v0.1.0', env: 'tst', commit: 'abc1234' });
  });

  it('mostra os dois commits so quando de fato diferem', () => {
    const parts = badgeParts(
      { version: '0.1.0', commit: 'abc1234' },
      { version: '0.1.0', commit: 'fffffff0000', env: 'tst' },
    );
    expect(parts.commit).toBe('abc1234/fffffff');
  });

  it('sem release (a versao e so o hash) mostra v0.0.0 e mantem o commit', () => {
    const parts = badgeParts(
      { version: 'abc1234', commit: 'abc1234' },
      { version: 'abc1234', commit: FULL, env: 'tst' },
    );
    expect(parts).toEqual({ kind: 'version', web: 'v0.0.0', api: 'v0.0.0', env: 'tst', commit: 'abc1234' });
  });

  it('um lado em release e o outro nao', () => {
    const parts = badgeParts({ version: 'abc1234', commit: 'abc1234' }, { version: '0.2.0', commit: FULL, env: 'tst' });
    expect(parts).toMatchObject({ web: 'v0.0.0', api: 'v0.2.0' });
  });

  it('sem resposta da API mostra so o web', () => {
    expect(badgeParts({ version: '0.1.0', commit: FULL })).toEqual({
      kind: 'versionWebOnly',
      web: 'v0.1.0',
      commit: 'abc1234',
    });
    expect(badgeParts({ version: 'dev', commit: 'local' }).web).toBe('v0.0.0');
  });

  it('versao como 0.1.0-5-gabc1234 (git describe) tambem e release', () => {
    expect(badgeParts({ version: '0.1.0-5-gabc1234', commit: 'abc1234' }).web).toBe('v0.1.0-5-gabc1234');
  });
});
