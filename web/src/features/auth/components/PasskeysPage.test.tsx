import { cleanup, fireEvent, screen, within } from '@testing-library/react';
import { afterEach, beforeAll, describe, expect, it, vi } from 'vitest';
import { json, me, renderApp, setupTestApp, status, stubApi, stubWebAuthn, version } from '../../../test-support';

beforeAll(setupTestApp);

afterEach(() => {
  cleanup();
  vi.unstubAllGlobals();
});

const first = {
  id: '11111111-1111-4111-8111-111111111111',
  name: 'Meu celular',
  created_at: '2026-10-01T12:00:00Z',
  last_used_at: '2026-10-10T12:00:00Z',
  backup_eligible: true,
};
const second = {
  id: '22222222-2222-4222-8222-222222222222',
  name: 'Notebook',
  created_at: '2026-10-05T12:00:00Z',
  backup_eligible: false,
};

function passkeysApi(list: unknown[], extra: Record<string, () => Response> = {}) {
  return stubApi({
    'GET /version': () => json(version),
    'GET /me': () => json(me),
    'GET /me/passkeys': () => json(list),
    ...extra,
  });
}

describe('PasskeysPage: lista e apagar', () => {
  it('sem sessão pede para entrar e leva ao início', async () => {
    stubApi({ 'GET /version': () => json(version), 'GET /me': () => status(401) });
    renderApp('/passkeys');
    expect(await screen.findByText('Entre para ver as suas passkeys.')).toBeInTheDocument();
    expect(screen.getByRole('link', { name: 'Voltar ao início' })).toHaveAttribute('href', '/');
  });

  it('falha ao carregar a lista mostra o aviso', async () => {
    passkeysApi([], { 'GET /me/passkeys': () => status(500) });
    renderApp('/passkeys');
    expect(await screen.findByRole('alert')).toHaveTextContent('Não foi possível carregar');
  });

  it('lista as passkeys com datas e "ainda não usada"', async () => {
    passkeysApi([first, second]);
    renderApp('/passkeys');
    expect(await screen.findByText('Meu celular')).toBeInTheDocument();
    expect(screen.getByText('Notebook')).toBeInTheDocument();
    expect(screen.getByText(/Último uso:/)).toBeInTheDocument();
    expect(screen.getByText('Ainda não usada')).toBeInTheDocument();
    expect(screen.getAllByRole('button', { name: 'Apagar' })).toHaveLength(2);
  });

  it('apagar chama a API com o cabeçalho anti-CSRF e recarrega a lista', async () => {
    let items: unknown[] = [first, second];
    const calls = passkeysApi([], {
      'GET /me/passkeys': () => json(items),
      [`DELETE /me/passkeys/${second.id}`]: () => {
        items = [first];
        return status(204);
      },
    });
    renderApp('/passkeys');
    const row = (await screen.findByText('Notebook')).closest('li');
    fireEvent.click(within(row as HTMLElement).getByRole('button', { name: 'Apagar' }));
    await vi.waitFor(() => expect(screen.queryByText('Notebook')).not.toBeInTheDocument());
    const del = calls.find((c) => c.method === 'DELETE');
    expect(del?.path).toBe(`/me/passkeys/${second.id}`);
    expect(del?.headers.get('X-Requested-With')).toBe('XMLHttpRequest');
  });

  it('a API recusar apagar a última (409) mostra o motivo e mantém a passkey', async () => {
    passkeysApi([first], { [`DELETE /me/passkeys/${first.id}`]: () => status(409) });
    renderApp('/passkeys');
    fireEvent.click(await screen.findByRole('button', { name: 'Apagar' }));
    expect(await screen.findByRole('alert')).toHaveTextContent('única forma de entrar');
    expect(screen.getByText('Meu celular')).toBeInTheDocument();
  });

  it('erro genérico ao apagar mostra aviso', async () => {
    passkeysApi([first, second], { [`DELETE /me/passkeys/${first.id}`]: () => status(500) });
    renderApp('/passkeys');
    const row = (await screen.findByText('Meu celular')).closest('li');
    fireEvent.click(within(row as HTMLElement).getByRole('button', { name: 'Apagar' }));
    expect(await screen.findByRole('alert')).toHaveTextContent('Não foi possível apagar');
  });
});

describe('PasskeysPage: adicionar', () => {
  it('adicionar registra outra passkey SEM convite (usa a sessão) e atualiza a lista', async () => {
    let items: unknown[] = [first];
    const calls = passkeysApi([], {
      'GET /me/passkeys': () => json(items),
      'POST /auth/passkeys/register/begin': () => json({ options: { challenge: 'abc' } }),
      'POST /auth/passkeys/register/finish': () => {
        items = [first, second];
        return json({ ...second }, 201);
      },
    });
    const credentials = stubWebAuthn();
    renderApp('/passkeys');
    fireEvent.click(await screen.findByRole('button', { name: 'Adicionar passkey' }));
    expect(await screen.findByText('Notebook')).toBeInTheDocument();
    expect(credentials.create).toHaveBeenCalledTimes(1);
    expect(calls.find((c) => c.path === '/auth/passkeys/register/begin')?.body).toEqual({});
  });

  it('passkey duplicada (409 ou aparelho que já tem) mostra aviso próprio', async () => {
    passkeysApi([first], { 'POST /auth/passkeys/register/begin': () => json({ options: { challenge: 'abc' } }) });
    stubWebAuthn({ create: async () => Promise.reject(new DOMException('x', 'InvalidStateError')) });
    renderApp('/passkeys');
    fireEvent.click(await screen.findByRole('button', { name: 'Adicionar passkey' }));
    expect(await screen.findByRole('alert')).toHaveTextContent('já está cadastrada');
  });

  it('conta sem passkeys mostra o texto vazio', async () => {
    passkeysApi([]);
    renderApp('/passkeys');
    expect(await screen.findByText('Nenhuma passkey ainda.')).toBeInTheDocument();
  });
});
