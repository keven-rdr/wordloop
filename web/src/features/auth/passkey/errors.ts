import { PasskeyError, type PasskeyFailure } from './browser';

export type PasskeyMessageKey =
  | 'passkeyCancelled'
  | 'passkeyUnsupported'
  | 'passkeyDuplicate'
  | 'passkeyUnauthorized'
  | 'passkeyTooMany'
  | 'passkeyFailed';

const keys: Record<PasskeyFailure, PasskeyMessageKey> = {
  cancelled: 'passkeyCancelled',
  unsupported: 'passkeyUnsupported',
  duplicate: 'passkeyDuplicate',
  unauthorized: 'passkeyUnauthorized',
  tooMany: 'passkeyTooMany',
  failed: 'passkeyFailed',
};

export function passkeyMessageKey(error: unknown): PasskeyMessageKey {
  return error instanceof PasskeyError ? keys[error.kind] : 'passkeyFailed';
}

export function ensureOk(response: Response | undefined): void {
  if (!response) throw new PasskeyError('failed');
  if (response.ok) return;
  if (response.status === 401) throw new PasskeyError('unauthorized');
  if (response.status === 409) throw new PasskeyError('duplicate');
  if (response.status === 429) throw new PasskeyError('tooMany');
  throw new PasskeyError('failed');
}
