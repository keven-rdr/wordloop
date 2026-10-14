export type PasswordFailure = 'invalid' | 'twoFactor' | 'weak' | 'inviteInvalid' | 'tooMany' | 'failed';

export class PasswordError extends Error {
  constructor(readonly kind: PasswordFailure) {
    super(kind);
  }
}

export type PasswordMessageKey =
  | 'password.invalid'
  | 'password.twoFactor'
  | 'password.weak'
  | 'password.inviteInvalid'
  | 'password.tooMany'
  | 'password.failed';

const keys: Record<PasswordFailure, PasswordMessageKey> = {
  invalid: 'password.invalid',
  twoFactor: 'password.twoFactor',
  weak: 'password.weak',
  inviteInvalid: 'password.inviteInvalid',
  tooMany: 'password.tooMany',
  failed: 'password.failed',
};

export function passwordMessageKey(error: unknown): PasswordMessageKey {
  return error instanceof PasswordError ? keys[error.kind] : 'password.failed';
}

const statusToFailure: Record<'login' | 'set', Record<number, PasswordFailure>> = {
  login: { 401: 'invalid', 403: 'twoFactor', 429: 'tooMany' },
  set: { 401: 'inviteInvalid', 422: 'weak', 429: 'tooMany' },
};

export function ensurePasswordOk(response: Response | undefined, flow: 'login' | 'set'): void {
  if (!response) throw new PasswordError('failed');
  if (response.ok) return;
  throw new PasswordError(statusToFailure[flow][response.status] ?? 'failed');
}
