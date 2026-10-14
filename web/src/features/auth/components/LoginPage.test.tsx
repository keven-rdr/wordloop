import { cleanup, fireEvent, screen } from '@testing-library/react';
import { afterEach, beforeAll, describe, expect, it, vi } from 'vitest';
import {
  json,
  me,
  renderApp,
  setupTestApp,
  status,
  stubApi,
  stubWebAuthn,
  version,
  withoutWebAuthn,
} from '../../../test-support';

beforeAll(setupTestApp);

afterEach(() => {
  cleanup();
  vi.unstubAllGlobals();
});

const options = { options: { challenge: 'abc', rpId: 'localhost' } };

function loginApi(extra: Record<string, () => Response> = {}) {
  let signedIn = false;
  return stubApi({
    'GET /version': () => json(version),
    'GET /me': () => (signedIn ? json(me) : status(401)),
    'POST /auth/passkeys/login/begin': () => json(options),
    'POST /auth/passkeys/login/finish': () => {
      signedIn = true;
      return status(204);
    },
    'POST /auth/logout': () => {
      signedIn = false;
      return status(204);
    },
    ...extra,
  });
}

describe('LoginPage: entrar', () => {
  it('sem sessão mostra o botão Entrar com passkey (nada de link para provedor externo)', async () => {
    loginApi();
    stubWebAuthn();
    renderApp('/');
    expect(await screen.findByRole('button', { name: 'Entrar com passkey' })).toBeInTheDocument();
    expect(screen.queryByRole('link', { name: /entrar/i })).not.toBeInTheDocument();
  });

  it('entrar: pede o desafio, usa a passkey do aparelho, conclui e mostra Olá', async () => {
    const calls = loginApi();
    const credentials = stubWebAuthn();
    renderApp('/');
    fireEvent.click(await screen.findByRole('button', { name: 'Entrar com passkey' }));
    expect(await screen.findByText('Olá, Maria!')).toBeInTheDocument();
    expect(credentials.get).toHaveBeenCalledTimes(1);
    const begin = calls.find((c) => c.path === '/auth/passkeys/login/begin');
    const finish = calls.find((c) => c.path === '/auth/passkeys/login/finish');
    expect(begin?.headers.get('X-Requested-With')).toBe('XMLHttpRequest');
    expect(finish?.headers.get('X-Requested-With')).toBe('XMLHttpRequest');
    expect(finish?.body).toEqual({ credential: { id: 'cred', type: 'public-key' } });
  });

  it('cancelar o pedido do aparelho mostra um aviso calmo e mantém o botão', async () => {
    loginApi();
    stubWebAuthn({ get: async () => Promise.reject(new DOMException('cancelado', 'NotAllowedError')) });
    renderApp('/');
    fireEvent.click(await screen.findByRole('button', { name: 'Entrar com passkey' }));
    expect(await screen.findByRole('status')).toHaveTextContent('Pedido cancelado');
    expect(screen.queryByRole('alert')).not.toBeInTheDocument();
    expect(screen.getByRole('button', { name: 'Entrar com passkey' })).toBeEnabled();
  });

  it('navegador sem passkeys mostra o aviso e não chama a API de conclusão', async () => {
    const calls = loginApi();
    withoutWebAuthn();
    renderApp('/');
    fireEvent.click(await screen.findByRole('button', { name: 'Entrar com passkey' }));
    expect(await screen.findByRole('alert')).toHaveTextContent('não suporta passkeys');
    expect(calls.some((c) => c.path.endsWith('/finish'))).toBe(false);
  });

  it('passkey não reconhecida (401) e excesso de tentativas (429) têm mensagens próprias', async () => {
    let answer = 401;
    loginApi({ 'POST /auth/passkeys/login/finish': () => status(answer) });
    stubWebAuthn();
    renderApp('/');
    fireEvent.click(await screen.findByRole('button', { name: 'Entrar com passkey' }));
    expect(await screen.findByRole('alert')).toHaveTextContent('não reconhecido');
    answer = 429;
    fireEvent.click(screen.getByRole('button', { name: 'Entrar com passkey' }));
    expect(await screen.findByText(/Muitas tentativas/)).toBeInTheDocument();
  });
});

describe('Início: com sessão', () => {
  it('com sessão mostra Olá, o atalho para as passkeys e Sair', async () => {
    loginApi({ 'GET /me': () => json(me) });
    stubWebAuthn();
    renderApp('/');
    expect(await screen.findByText('Olá, Maria!')).toBeInTheDocument();
    expect(screen.getByRole('link', { name: 'Minhas passkeys' })).toHaveAttribute('href', '/passkeys');
    expect(screen.getByRole('button', { name: 'Sair' })).toBeInTheDocument();
    expect(screen.queryByRole('button', { name: 'Entrar com passkey' })).not.toBeInTheDocument();
  });

  it('Sair envia o cabeçalho anti-CSRF e volta para Entrar com passkey', async () => {
    const calls = loginApi();
    stubWebAuthn();
    renderApp('/');
    fireEvent.click(await screen.findByRole('button', { name: 'Entrar com passkey' }));
    fireEvent.click(await screen.findByRole('button', { name: 'Sair' }));
    expect(await screen.findByRole('button', { name: 'Entrar com passkey' })).toBeInTheDocument();
    const logout = calls.find((c) => c.path === '/auth/logout');
    expect(logout?.method).toBe('POST');
    expect(logout?.headers.get('X-Requested-With')).toBe('XMLHttpRequest');
  });

  it('falha ao sair mostra aviso e mantém a pessoa logada', async () => {
    loginApi({ 'GET /me': () => json(me), 'POST /auth/logout': () => status(500) });
    stubWebAuthn();
    renderApp('/');
    fireEvent.click(await screen.findByRole('button', { name: 'Sair' }));
    expect(await screen.findByRole('alert')).toHaveTextContent('Não foi possível sair');
    expect(screen.getByText('Olá, Maria!')).toBeInTheDocument();
  });

  it('enquanto carrega não mostra nenhum botão', () => {
    loginApi();
    stubWebAuthn();
    renderApp('/');
    expect(screen.queryByRole('button', { name: 'Entrar com passkey' })).not.toBeInTheDocument();
    expect(screen.queryByRole('button', { name: 'Sair' })).not.toBeInTheDocument();
  });
});

describe('LoginPage: estados', () => {
  it('enquanto a passkey é verificada mostra "Verificando…" e desabilita o botão', async () => {
    let release: (value: Response) => void = () => undefined;
    stubApi({
      'GET /version': () => json(version),
      'GET /me': () => status(401),
      'POST /auth/passkeys/login/begin': () =>
        new Promise<Response>((resolve) => {
          release = resolve;
        }),
    });
    stubWebAuthn();
    renderApp('/login');
    fireEvent.click(await screen.findByRole('button', { name: 'Entrar com passkey' }));
    expect(await screen.findByRole('button', { name: 'Verificando…' })).toBeDisabled();
    release(status(500));
  });

  it('sem nome no perfil o início cumprimenta só com "Olá!"', async () => {
    loginApi({ 'GET /me': () => json({ ...me, display_name: '' }) });
    renderApp('/');
    expect(await screen.findByText('Olá!')).toBeInTheDocument();
  });
});
