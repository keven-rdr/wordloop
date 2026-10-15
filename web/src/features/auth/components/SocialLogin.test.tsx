import { cleanup, screen } from '@testing-library/react';
import { afterEach, beforeAll, describe, expect, it, vi } from 'vitest';
import { json, renderApp, setupTestApp, status, stubApi, version } from '../../../test-support';

beforeAll(setupTestApp);

afterEach(() => {
  cleanup();
  vi.unstubAllGlobals();
});

function providersApi(providers: string[] | undefined) {
  stubApi({
    'GET /version': () => json(version),
    'GET /me': () => status(401),
    ...(providers ? { 'GET /auth/providers': () => json({ providers }) } : {}),
  });
}

describe('Botões de login social', () => {
  it.each(['/login', '/signup'])('%s: com Google e GitHub ligados os dois são links para o servidor', async (path) => {
    providersApi(['google', 'github']);
    renderApp(path);
    const google = await screen.findByRole('link', { name: /Continuar com Google/ });
    const github = await screen.findByRole('link', { name: /Continuar com GitHub/ });
    expect(google).toHaveAttribute('href', '/api/v1/auth/login?provider=google');
    expect(github).toHaveAttribute('href', '/api/v1/auth/login?provider=github');
    expect(screen.queryByText('em breve')).not.toBeInTheDocument();
  });

  it('só o provedor ligado vira link; o outro fica desligado com "em breve"', async () => {
    providersApi(['github']);
    renderApp('/login');
    expect(await screen.findByRole('link', { name: /Continuar com GitHub/ })).toBeInTheDocument();
    const google = await screen.findByRole('button', { name: /Continuar com Google/ });
    expect(google).toBeDisabled();
    expect(google).toHaveTextContent('em breve');
  });

  it.each([[[]], [undefined]])('sem provedores (%j) os dois ficam desligados', async (providers) => {
    providersApi(providers);
    renderApp('/login');
    expect(await screen.findByRole('button', { name: /Continuar com Google/ })).toBeDisabled();
    expect(await screen.findByRole('button', { name: /Continuar com GitHub/ })).toBeDisabled();
    expect(screen.queryByRole('link', { name: /Continuar com/ })).not.toBeInTheDocument();
  });
});

describe('Enquanto a lista de provedores carrega', () => {
  it('os botões não mostram "em breve" antes de a resposta chegar', async () => {
    stubApi({
      'GET /version': () => json(version),
      'GET /me': () => status(401),
      'GET /auth/providers': () => new Promise<Response>(() => undefined) as unknown as Response,
    });
    renderApp('/login');
    expect(await screen.findByRole('button', { name: /Continuar com Google/ })).toBeDisabled();
    expect(screen.queryByText('em breve')).not.toBeInTheDocument();
  });
});

describe('Volta do provedor com erro', () => {
  it.each([
    ['failed', 'Não foi possível entrar com o provedor'],
    ['two_factor', 'exige 2 fatores'],
    ['email', 'não informou um e-mail confirmado'],
  ])('/login?error=%s mostra a mensagem certa', async (code, message) => {
    providersApi(['google']);
    renderApp(`/login?error=${code}`);
    expect(await screen.findByRole('alert')).toHaveTextContent(message);
  });

  it('um código de erro desconhecido na URL é ignorado', async () => {
    providersApi(['google']);
    renderApp('/login?error=<script>');
    expect(await screen.findByRole('heading', { name: 'Entrar' })).toBeInTheDocument();
    expect(screen.queryByRole('alert')).not.toBeInTheDocument();
  });
});
