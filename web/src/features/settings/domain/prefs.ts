import type { Settings } from '../../../api/generated/types.gen';

export type UiLocale = Settings['ui_locale'];
export type Theme = Settings['theme'];

export const LOCALES: readonly UiLocale[] = ['system', 'pt-BR', 'en-US'];
export const THEMES: readonly Theme[] = ['system', 'light', 'dark'];

// O fuso que o aparelho informa (nome IANA, como America/Sao_Paulo).
export const detectTimeZone = () => Intl.DateTimeFormat().resolvedOptions().timeZone;

type WithZones = { supportedValuesOf?: (key: 'timeZone') => string[] };

// Os fusos para escolher: os que o navegador conhece, mais o atual e o do aparelho (sem repetir, em ordem).
export function timeZones(current: string, device: string): string[] {
  const known = (Intl as unknown as WithZones).supportedValuesOf?.('timeZone') ?? [];
  return [...new Set([...known, current, device].filter(Boolean))].sort((a, b) => a.localeCompare(b));
}

// O fuso passa a acompanhar o aparelho.
export const followDevice = (settings: Settings, device: string): Settings => ({
  ...settings,
  tz: device,
  tz_auto: true,
});

// A pessoa escolheu um fuso: fica fixo e o aparelho não o altera mais.
export const pinZone = (settings: Settings, tz: string): Settings => ({ ...settings, tz, tz_auto: false });

// O idioma que vale na interface: o escolhido ou, em "sistema", o do aparelho (pt para português, en para inglês).
export function targetLocale(choice: UiLocale, languages: readonly string[]): 'pt-BR' | 'en-US' {
  if (choice !== 'system') return choice;
  for (const language of languages) {
    const base = language.toLowerCase().split('-')[0];
    if (base === 'pt') return 'pt-BR';
    if (base === 'en') return 'en-US';
  }
  return 'pt-BR';
}

export type Appliers = Readonly<{
  language: string;
  languages: readonly string[];
  changeLanguage: (locale: string) => void;
  setTheme: (theme: Theme) => void;
}>;

// Põe na tela o que está salvo: o tema e, se mudou, o idioma. Só mexe no que é diferente do que já vale.
export function applyPreferences(settings: Settings, { language, languages, changeLanguage, setTheme }: Appliers) {
  setTheme(settings.theme);
  const locale = targetLocale(settings.ui_locale, languages);
  if (locale !== language) changeLanguage(locale);
}
