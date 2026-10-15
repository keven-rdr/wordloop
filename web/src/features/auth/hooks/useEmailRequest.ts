import { useMutation } from '@tanstack/react-query';
import { useTranslation } from 'react-i18next';
import { registerAccount, requestPasswordReset } from '../../../api/generated/sdk.gen';
import { ensureEmailRequestOk } from '../email/errors';
import { csrfHeaders } from './csrf';

export type EmailFlow = 'signup' | 'forgot';

const supported = ['pt-BR', 'en-US'] as const;

function useLocale() {
  const { i18n } = useTranslation();
  return supported.find((locale) => locale === i18n.language);
}

export function useEmailRequest(flow: EmailFlow) {
  const locale = useLocale();
  return useMutation({
    mutationFn: async (email: string) => {
      const send = flow === 'signup' ? registerAccount : requestPasswordReset;
      const { response } = await send({ headers: csrfHeaders, body: { email, locale } });
      ensureEmailRequestOk(response);
      return email;
    },
  });
}
