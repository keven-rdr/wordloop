export type PasskeyFailure = 'unsupported' | 'cancelled' | 'duplicate' | 'unauthorized' | 'tooMany' | 'failed';

export class PasskeyError extends Error {
  constructor(readonly kind: PasskeyFailure) {
    super(kind);
  }
}

type JsonCredential = { toJSON(): Record<string, unknown> };

type Statics = {
  parseCreationOptionsFromJSON?: (options: unknown) => PublicKeyCredentialCreationOptions;
  parseRequestOptionsFromJSON?: (options: unknown) => PublicKeyCredentialRequestOptions;
};

function statics(): Statics | undefined {
  return typeof PublicKeyCredential === 'undefined' ? undefined : (PublicKeyCredential as unknown as Statics);
}

export function supportsPasskeys(): boolean {
  const s = statics();
  return Boolean(s?.parseCreationOptionsFromJSON && s.parseRequestOptionsFromJSON && navigator.credentials);
}

function toFailure(error: unknown): PasskeyError {
  if (error instanceof PasskeyError) return error;
  const name = error instanceof DOMException ? error.name : '';
  if (name === 'NotAllowedError' || name === 'AbortError') return new PasskeyError('cancelled');
  if (name === 'InvalidStateError') return new PasskeyError('duplicate');
  return new PasskeyError('failed');
}

export async function createCredential(options: unknown): Promise<Record<string, unknown>> {
  const s = statics();
  if (!supportsPasskeys() || !s?.parseCreationOptionsFromJSON) throw new PasskeyError('unsupported');
  try {
    const publicKey = s.parseCreationOptionsFromJSON(options);
    const credential = (await navigator.credentials.create({ publicKey })) as unknown as JsonCredential | null;
    if (!credential) throw new PasskeyError('failed');
    return credential.toJSON();
  } catch (error) {
    throw toFailure(error);
  }
}

export async function getCredential(options: unknown): Promise<Record<string, unknown>> {
  const s = statics();
  if (!supportsPasskeys() || !s?.parseRequestOptionsFromJSON) throw new PasskeyError('unsupported');
  try {
    const publicKey = s.parseRequestOptionsFromJSON(options);
    const credential = (await navigator.credentials.get({ publicKey })) as unknown as JsonCredential | null;
    if (!credential) throw new PasskeyError('failed');
    return credential.toJSON();
  } catch (error) {
    throw toFailure(error);
  }
}
