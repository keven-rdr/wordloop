import { describe, expect, it } from 'vitest';
import { badgeParts } from './formatVersion';

describe('badgeParts', () => {
  it('web e API na mesma versao mostram so uma', () => {
    expect(badgeParts('1.0.162', '1.0.162')).toEqual({ kind: 'version', web: 'v1.0.162' });
  });

  it('versoes diferentes mostram as duas, para ver o que ficou para tras', () => {
    expect(badgeParts('1.0.162', '1.0.150')).toEqual({ kind: 'versionWithApi', web: 'v1.0.162', api: 'v1.0.150' });
  });

  it('sem resposta da API mostra so o web', () => {
    expect(badgeParts('1.0.162')).toEqual({ kind: 'version', web: 'v1.0.162' });
  });

  it('fora do CI (dev local) fica como veio, sem o v', () => {
    expect(badgeParts('dev', 'dev')).toEqual({ kind: 'version', web: 'dev' });
    expect(badgeParts('dev', '1.0.5')).toEqual({ kind: 'versionWithApi', web: 'dev', api: 'v1.0.5' });
  });

  it('versao como 1.0.162-3-gabc (git describe) tambem vira v', () => {
    expect(badgeParts('1.0.162-3-gabc1234').web).toBe('v1.0.162-3-gabc1234');
  });
});
