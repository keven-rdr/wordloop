import { cleanup, fireEvent, screen } from '@testing-library/react';
import { afterEach, beforeAll, describe, expect, it, vi } from 'vitest';
import { json, me, renderApp, setupTestApp, status, stubApi, stubWebAuthn, version } from '../../../test-support';

beforeAll(setupTestApp);

afterEach(() => {
  cleanup();
  vi.unstubAllGlobals();
  window.history.replaceState(null, '', '/');
});

const options = { options: { challenge: 'abc', rp: { id: 'localhost' } } };

function inviteApi(extra: Record<string, () => Response> = {}) {
  let signedIn = false;
  return stubApi({
    'GET /version': () => json(version),
    'GET /me': () => (signedIn ? json(me) : status(401)),
    'POST /auth/passkeys/register/begin': () => json(options),
    'POST /auth/passkeys/register/finish': () => {
      signedIn = true;
      return json({ id: me.id, name: 'Passkey', created_at: '2026-10-15T12:00:00Z', backup_eligible: false }, 201);
    },
    ...extra,
  });
}

function openInvite(secret: string) {
  window.history.replaceState(null, '', `/invite${secret ? `#${secret}` : ''}`);
  return renderApp('/invite');
}

describe('InvitePage', () => {
  it('lê o segredo do fragmento e o apaga da barra de endereço logo ao abrir', async () => {
    inviteApi();
    stubWebAuthn();
    openInvite('segredo-do-convite');
    expect(await screen.findByRole('button', { name: 'Criar minha passkey' })).toBeInTheDocument();
    expect(window.location.hash).toBe('');
    expect(window.location.href).not.toContain('segredo-do-convite');
  });

  it('criar a passkey manda o convite no corpo, usa o aparelho, conclui e vai para o início logado', async () => {
    const calls = inviteApi();
    const credentials = stubWebAuthn();
    openInvite('segredo-do-convite');
    fireEvent.click(await screen.findByRole('button', { name: 'Criar minha passkey' }));
    expect(await screen.findByText('Olá, Maria!')).toBeInTheDocument();
    expect(credentials.create).toHaveBeenCalledTimes(1);
    const begin = calls.find((c) => c.path === '/auth/passkeys/register/begin');
    const finish = calls.find((c) => c.path === '/auth/passkeys/register/finish');
    expect(begin?.body).toEqual({ invite: 'segredo-do-convite' });
    expect(begin?.headers.get('X-Requested-With')).toBe('XMLHttpRequest');
    expect(finish?.body).toMatchObject({ credential: { id: 'cred', type: 'public-key' } });
  });

  it('sem segredo no link mostra que o convite não é válido e não chama a API', async () => {
    const calls = inviteApi();
    stubWebAuthn();
    openInvite('');
    expect(await screen.findByRole('alert')).toHaveTextContent('não é válido ou já foi usado');
    expect(screen.queryByRole('button', { name: 'Criar minha passkey' })).not.toBeInTheDocument();
    expect(calls.some((c) => c.path.includes('/passkeys/'))).toBe(false);
  });

  it('convite recusado pela API (usado ou vencido) mostra o aviso e permite tentar de novo', async () => {
    inviteApi({ 'POST /auth/passkeys/register/begin': () => status(401) });
    stubWebAuthn();
    openInvite('segredo-vencido');
    fireEvent.click(await screen.findByRole('button', { name: 'Criar minha passkey' }));
    expect(await screen.findByRole('alert')).toHaveTextContent('15 minutos e uma vez só');
    expect(screen.getByRole('button', { name: 'Criar minha passkey' })).toBeEnabled();
  });

  it('o aparelho recusar (cancelar) não consome o convite: nenhuma chamada de conclusão', async () => {
    const calls = inviteApi();
    stubWebAuthn({ create: async () => Promise.reject(new DOMException('x', 'NotAllowedError')) });
    openInvite('segredo');
    fireEvent.click(await screen.findByRole('button', { name: 'Criar minha passkey' }));
    expect(await screen.findByRole('alert')).toHaveTextContent('Pedido cancelado');
    expect(calls.some((c) => c.path.endsWith('/register/finish'))).toBe(false);
  });
});
