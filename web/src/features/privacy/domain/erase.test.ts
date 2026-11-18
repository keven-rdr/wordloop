import { describe, expect, it } from 'vitest';
import { canConfirm, EraseError, exportFileName, failureOf, MESSAGE_KEYS } from './erase';

describe('failureOf', () => {
  it.each([
    [204, undefined],
    [400, 'mismatch'],
    [401, 'session'],
    [403, 'admin'],
    [428, 'recent'],
    [500, 'failed'],
    [503, 'failed'],
    [404, 'failed'],
  ] as const)('status %i → %s', (status, expected) => {
    expect(failureOf(status)).toBe(expected);
  });

  it('todo motivo tem uma mensagem', () => {
    for (const failure of ['mismatch', 'admin', 'recent', 'session', 'failed'] as const) {
      expect(MESSAGE_KEYS[failure]).toBe(`erase.errors.${failure}`);
    }
  });

  it('o erro guarda o motivo', () => {
    expect(new EraseError('recent').failure).toBe('recent');
  });
});

describe('exportFileName', () => {
  it('usa a data em AAAA-MM-DD', () => {
    expect(exportFileName(new Date('2026-11-18T23:59:00Z'))).toBe('wordloop-dados-2026-11-18.json');
  });
});

describe('canConfirm', () => {
  it('só com alguma coisa digitada', () => {
    expect(canConfirm('')).toBe(false);
    expect(canConfirm('   ')).toBe(false);
    expect(canConfirm('ana@example.com')).toBe(true);
    expect(canConfirm(' a ')).toBe(true);
  });
});
