import { cleanup, fireEvent, screen, waitFor } from '@testing-library/react';
import { afterEach, beforeAll, describe, expect, it, vi } from 'vitest';
import { json, me, renderApp, setupTestApp, status, stubApi, version } from '../../../test-support';

beforeAll(setupTestApp);

afterEach(() => {
  cleanup();
  vi.unstubAllGlobals();
});

function api(extra: Record<string, () => Response> = {}) {
  let signedIn = true;
  return stubApi({
    'GET /version': () => json(version),
    'GET /me': () => (signedIn ? json(me) : status(401)),
    'GET /me/passkeys': () => json([]),
    'POST /auth/logout': () => {
      signedIn = false;
      return status(204);
    },
    ...extra,
  });
}

describe('Configurações', () => {
  it('mostra o nome, as passkeys, os créditos, o tema e Sair', async () => {
    api();
    renderApp('/settings');
    expect(await screen.findByRole('heading', { name: 'Configurações' })).toBeInTheDocument();
    expect(screen.getByText('Maria')).toBeInTheDocument();
    expect(screen.getByRole('link', { name: /Minhas passkeys/ })).toHaveAttribute('href', '/passkeys');
    expect(screen.getByRole('link', { name: /Créditos e licenças/ })).toHaveAttribute('href', '/words/credits');
    expect(screen.getByText('Tema claro ou escuro')).toBeInTheDocument();
    expect(screen.getByRole('button', { name: 'Sair' })).toBeInTheDocument();
  });

  it('mostra o selo da versão', async () => {
    api();
    renderApp('/settings');
    const badge = await screen.findByTestId('version-badge');
    await waitFor(() => expect(badge).toHaveTextContent('API v0.1.0'));
  });

  it('Minhas passkeys leva à lista e ela volta para as configurações', async () => {
    api();
    const { router } = renderApp('/settings');
    fireEvent.click(await screen.findByRole('link', { name: /Minhas passkeys/ }));
    await waitFor(() => expect(router.state.location.pathname).toBe('/passkeys'));
    fireEvent.click(await screen.findByRole('link', { name: 'Voltar às configurações' }));
    await waitFor(() => expect(router.state.location.pathname).toBe('/settings'));
  });

  it('Sair encerra a sessão e leva ao login', async () => {
    const calls = api();
    renderApp('/settings');
    fireEvent.click(await screen.findByRole('button', { name: 'Sair' }));
    expect(await screen.findByRole('heading', { name: 'Entrar' })).toBeInTheDocument();
    expect(calls.find((c) => c.path === '/auth/logout')?.headers.get('X-Requested-With')).toBe('XMLHttpRequest');
  });

  it('sem sessão volta para o login', async () => {
    api({ 'GET /me': () => status(401) });
    renderApp('/settings');
    expect(await screen.findByRole('heading', { name: 'Entrar' })).toBeInTheDocument();
  });

  it('sem nome mostra a saudação neutra', async () => {
    api({ 'GET /me': () => json({ ...me, display_name: '' }) });
    renderApp('/settings');
    expect(await screen.findByText('Olá!')).toBeInTheDocument();
  });
});
