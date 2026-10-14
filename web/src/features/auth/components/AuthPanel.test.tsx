import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { cleanup, fireEvent, render, screen } from '@testing-library/react';
import { afterEach, beforeAll, describe, expect, it, vi } from 'vitest';
import { configureApi } from '../../../core/api';
import { setupI18n } from '../../../core/i18n';
import { AuthPanel } from './AuthPanel';

beforeAll(async () => {
  configureApi('http://localhost');
  await setupI18n('pt-BR');
});

afterEach(() => {
  cleanup();
  vi.unstubAllGlobals();
});

const me = { id: '3f0c4f4e-8d1e-4a3b-9a55-1d2c3b4a5e6f', display_name: 'Maria', role: 'user' };

function renderPanel() {
  render(
    <QueryClientProvider client={new QueryClient({ defaultOptions: { queries: { retry: false } } })}>
      <AuthPanel />
    </QueryClientProvider>,
  );
}

type Handler = (req: Request) => Response;

function stubApi(handler: Handler) {
  const calls: Request[] = [];
  vi.stubGlobal(
    'fetch',
    vi.fn(async (input: Request | string, init?: RequestInit) => {
      const req = input instanceof Request ? input : new Request(input, init);
      calls.push(req);
      return handler(req);
    }),
  );
  return calls;
}

describe('AuthPanel', () => {
  it('sem sessão mostra o botão Entrar apontando para o login da API', async () => {
    stubApi(() => new Response(null, { status: 401 }));
    renderPanel();
    const link = await screen.findByRole('link', { name: 'Entrar' });
    expect(link).toHaveAttribute('href', '/api/v1/auth/login?return_to=/');
    expect(screen.queryByRole('button', { name: 'Sair' })).not.toBeInTheDocument();
  });

  it('com sessão mostra Olá com o nome e o botão Sair', async () => {
    stubApi(() => Response.json(me));
    renderPanel();
    expect(await screen.findByText('Olá, Maria!')).toBeInTheDocument();
    expect(screen.getByRole('button', { name: 'Sair' })).toBeInTheDocument();
    expect(screen.queryByRole('link', { name: 'Entrar' })).not.toBeInTheDocument();
  });

  it('Sair envia o cabeçalho anti-CSRF e volta para Entrar', async () => {
    let loggedIn = true;
    const calls = stubApi((req) => {
      if (req.url.endsWith('/auth/logout')) {
        loggedIn = false;
        return new Response(null, { status: 204 });
      }
      return loggedIn ? Response.json(me) : new Response(null, { status: 401 });
    });
    renderPanel();
    fireEvent.click(await screen.findByRole('button', { name: 'Sair' }));
    expect(await screen.findByRole('link', { name: 'Entrar' })).toBeInTheDocument();
    const logout = calls.find((c) => c.url.endsWith('/auth/logout'));
    expect(logout?.method).toBe('POST');
    expect(logout?.headers.get('X-Requested-With')).toBe('XMLHttpRequest');
  });

  it('falha ao sair mostra aviso e mantém a pessoa logada', async () => {
    stubApi((req) => (req.url.endsWith('/auth/logout') ? new Response(null, { status: 500 }) : Response.json(me)));
    renderPanel();
    fireEvent.click(await screen.findByRole('button', { name: 'Sair' }));
    expect(await screen.findByRole('alert')).toHaveTextContent('Não foi possível sair');
    expect(screen.getByText('Olá, Maria!')).toBeInTheDocument();
  });

  it('enquanto carrega não mostra nenhum botão', () => {
    stubApi(() => new Response(null, { status: 401 }));
    renderPanel();
    expect(screen.queryByRole('link', { name: 'Entrar' })).not.toBeInTheDocument();
    expect(screen.queryByRole('button', { name: 'Sair' })).not.toBeInTheDocument();
  });
});
