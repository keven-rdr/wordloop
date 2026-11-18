import { describe, expect, it, vi } from 'vitest';
import type { Settings } from '../../../api/generated/types.gen';
import { applyPreferences, followDevice, pinZone, targetLocale, timeZones } from './prefs';

const settings: Settings = { ui_locale: 'system', theme: 'system', tz: 'America/Sao_Paulo', tz_auto: true };

describe('fuso', () => {
  it('seguir o aparelho troca o fuso pelo do aparelho e liga o automático', () => {
    expect(followDevice({ ...settings, tz: 'Asia/Tokyo', tz_auto: false }, 'Europe/Lisbon')).toEqual({
      ...settings,
      tz: 'Europe/Lisbon',
      tz_auto: true,
    });
  });

  it('escolher um fuso o fixa e desliga o automático', () => {
    expect(pinZone(settings, 'Asia/Tokyo')).toEqual({ ...settings, tz: 'Asia/Tokyo', tz_auto: false });
  });

  it('a lista traz os fusos do navegador, o atual e o do aparelho, sem repetir e em ordem', () => {
    const list = timeZones('Etc/Fantasia', 'America/Sao_Paulo');
    expect(list).toContain('Etc/Fantasia');
    expect(list).toContain('America/Sao_Paulo');
    expect(new Set(list).size).toBe(list.length);
    expect([...list].sort((a, b) => a.localeCompare(b))).toEqual(list);
  });

  it('sem fuso atual nem do aparelho, não inclui texto vazio', () => {
    expect(timeZones('', '')).not.toContain('');
  });
});

describe('targetLocale', () => {
  it.each([
    ['pt-BR', ['en-US'], 'pt-BR'],
    ['en-US', ['pt-BR'], 'en-US'],
    ['system', ['pt-PT', 'en'], 'pt-BR'],
    ['system', ['en-GB'], 'en-US'],
    ['system', ['fr-FR', 'en-US'], 'en-US'],
    ['system', ['fr-FR'], 'pt-BR'],
    ['system', [], 'pt-BR'],
  ] as const)('%s com aparelho %j → %s', (choice, languages, expected) => {
    expect(targetLocale(choice, languages)).toBe(expected);
  });
});

describe('applyPreferences', () => {
  const appliers = () => ({
    language: 'pt-BR',
    languages: ['pt-BR'],
    changeLanguage: vi.fn(),
    setTheme: vi.fn(),
  });

  it('aplica o tema e troca o idioma quando é diferente', () => {
    const a = appliers();
    applyPreferences({ ...settings, theme: 'dark', ui_locale: 'en-US' }, a);
    expect(a.setTheme).toHaveBeenCalledWith('dark');
    expect(a.changeLanguage).toHaveBeenCalledWith('en-US');
  });

  it('não troca o idioma que já vale', () => {
    const a = appliers();
    applyPreferences({ ...settings, ui_locale: 'pt-BR' }, a);
    expect(a.changeLanguage).not.toHaveBeenCalled();
  });

  it('em "sistema" segue o aparelho', () => {
    const a = { ...appliers(), language: 'pt-BR', languages: ['en-US'] };
    applyPreferences(settings, a);
    expect(a.changeLanguage).toHaveBeenCalledWith('en-US');
  });
});
