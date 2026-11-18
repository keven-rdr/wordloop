import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import i18n from 'i18next';
import {
  getSettingsOptions,
  getSettingsQueryKey,
  updateSettingsMutation,
} from '../../../api/generated/@tanstack/react-query.gen';
import type { Settings } from '../../../api/generated/types.gen';
import { setThemeChoice } from '../../../core/theme/useTheme';
import { applyPreferences } from '../domain/prefs';

// O servidor só aceita escrita com este cabeçalho (proteção contra CSRF).
const csrfHeaders = { 'X-Requested-With': 'XMLHttpRequest' } as const;

// Põe na tela o idioma e o tema salvos (o fuso é do servidor).
export function show(settings: Settings) {
  applyPreferences(settings, {
    language: i18n.language,
    languages: navigator.languages,
    changeLanguage: (locale) => void i18n.changeLanguage(locale),
    setTheme: setThemeChoice,
  });
}

// As preferências da pessoa. Sem sessão (401) não há dados e nada acontece.
export const useSettings = () => useQuery({ ...getSettingsOptions(), retry: false, refetchOnWindowFocus: false });

// Salva as preferências inteiras; ao voltar, guarda no cache e já aplica na tela.
export function useSaveSettings() {
  const queryClient = useQueryClient();
  const mutation = useMutation({
    ...updateSettingsMutation(),
    networkMode: 'always', // sem rede falha na hora e avisa, em vez de ficar pausada
    onSuccess: (saved) => {
      queryClient.setQueryData(getSettingsQueryKey(), saved);
      show(saved);
    },
  });
  const save = (body: Settings) => mutation.mutate({ headers: csrfHeaders, body });
  return { save, saving: mutation.isPending, failed: mutation.isError, saved: mutation.isSuccess };
}
