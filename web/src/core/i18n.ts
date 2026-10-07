import i18n from 'i18next';
import resourcesToBackend from 'i18next-resources-to-backend';
import { initReactI18next } from 'react-i18next';

export const locales = ['pt-BR', 'en-US'] as const;
export type Locale = (typeof locales)[number];

export function detectLocale(languages: readonly string[]): Locale {
  for (const lang of languages) {
    const base = lang.toLowerCase().split('-')[0];
    if (base === 'pt') return 'pt-BR';
    if (base === 'en') return 'en-US';
  }
  return 'pt-BR';
}

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
