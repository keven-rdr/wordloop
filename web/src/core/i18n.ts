import i18n from 'i18next';
import resourcesToBackend from 'i18next-resources-to-backend';
import { initReactI18next } from 'react-i18next';

export const locales = ['pt-BR', 'en-US'] as const;
export type Locale = (typeof locales)[number];

export function detectLocale(languages: readonly string[]): Locale {
  const match = languages.find((l) => l.toLowerCase().startsWith('en'));
  return match ? 'en-US' : 'pt-BR';
}

// Cada namespace e um JSON em src/locales/<idioma>/<ns>.json, carregado sob demanda (ADR 0015).
export function setupI18n(lng: Locale = detectLocale(navigator.languages)) {
  return i18n
    .use(initReactI18next)
    .use(resourcesToBackend((l: string, ns: string) => import(`../locales/${l}/${ns}.json`)))
    .init({
      lng,
      fallbackLng: 'pt-BR',
      supportedLngs: [...locales],
      ns: ['common'],
      defaultNS: 'common',
      interpolation: { escapeValue: false },
      react: { useSuspense: false },
    });
}
