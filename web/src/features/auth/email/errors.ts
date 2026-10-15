export type EmailFailure = 'invalid' | 'tooMany' | 'unavailable' | 'failed';

export class EmailError extends Error {
  constructor(readonly kind: EmailFailure) {
    super(kind);
  }
}

export type EmailMessageKey = 'email.invalid' | 'email.tooMany' | 'email.unavailable' | 'email.failed';

const keys: Record<EmailFailure, EmailMessageKey> = {
  invalid: 'email.invalid',
  tooMany: 'email.tooMany',
  unavailable: 'email.unavailable',
  failed: 'email.failed',
};

export function emailMessageKey(error: unknown): EmailMessageKey {
  return error instanceof EmailError ? keys[error.kind] : 'email.failed';
}

const statusToFailure: Record<number, EmailFailure> = { 400: 'invalid', 429: 'tooMany', 503: 'unavailable' };

export function ensureEmailRequestOk(response: Response | undefined): void {
  if (!response) throw new EmailError('failed');
  if (response.status === 202) return;
  throw new EmailError(statusToFailure[response.status] ?? 'failed');
}
