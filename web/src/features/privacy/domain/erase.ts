// Por que a exclusão não aconteceu, na linguagem da tela (o servidor responde só com o status).
export type EraseFailure = 'mismatch' | 'admin' | 'recent' | 'session' | 'failed';

export class EraseError extends Error {
  constructor(readonly failure: EraseFailure) {
    super(failure);
  }
}

// 204 apagou; 400 é o e-mail que não confere; 401 a sessão acabou; 403 conta de administrador; 428 login antigo.
export function failureOf(status: number): EraseFailure | undefined {
  switch (status) {
    case 204:
      return undefined;
    case 400:
      return 'mismatch';
    case 401:
      return 'session';
    case 403:
      return 'admin';
    case 428:
      return 'recent';
    default:
      return 'failed';
  }
}

export const MESSAGE_KEYS: Record<EraseFailure, string> = {
  mismatch: 'erase.errors.mismatch',
  admin: 'erase.errors.admin',
  recent: 'erase.errors.recent',
  session: 'erase.errors.session',
  failed: 'erase.errors.failed',
};

// O nome do arquivo da exportação: wordloop-dados-AAAA-MM-DD.json (a data do aparelho, em UTC, como o servidor).
export const exportFileName = (when: Date) => `wordloop-dados-${when.toISOString().slice(0, 10)}.json`;

// A pessoa só pode confirmar quando digitou alguma coisa: quem decide se é o e-mail certo é o servidor.
export const canConfirm = (typed: string) => typed.trim().length > 0;
