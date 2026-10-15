import { cleanup, fireEvent, screen, waitFor } from '@testing-library/react';
import { afterEach, beforeAll, describe, expect, it, vi } from 'vitest';
import { json, me, renderApp, setupTestApp, status, stubApi, version } from '../../../test-support';

beforeAll(setupTestApp);

afterEach(() => {
  cleanup();
  vi.unstubAllGlobals();
});

function anonymousApi() {
  stubApi({ 'GET /version': () => json(version), 'GET /me': () => status(401) });
}

describe('Login e cadastro: navegação', () => {
  it('o início sem sessão leva para /login', async () => {
    anonymousApi();
    const { router } = renderApp('/');
    expect(await screen.findByRole('heading', { name: 'Entrar' })).toBeInTheDocument();
    expect(router.state.location.pathname).toBe('/login');
  });

  it('de /login vai para /signup e volta', async () => {
    anonymousApi();
    const { router } = renderApp('/login');
    fireEvent.click(await screen.findByRole('link', { name: 'Criar conta' }));
    expect(await screen.findByRole('heading', { name: 'Criar conta' })).toBeInTheDocument();
    expect(router.state.location.pathname).toBe('/signup');
    fireEvent.click(screen.getByRole('link', { name: 'Entrar' }));
    expect(await screen.findByRole('heading', { name: 'Entrar' })).toBeInTheDocument();
  });

  it('quem já está logado é levado do /login e do /signup para o início', async () => {
    stubApi({ 'GET /version': () => json(version), 'GET /me': () => json(me) });
    const first = renderApp('/login');
    expect(await screen.findByText('Olá, Maria!')).toBeInTheDocument();
    expect(first.router.state.location.pathname).toBe('/');
    cleanup();
    const second = renderApp('/signup');
    expect(await screen.findByText('Olá, Maria!')).toBeInTheDocument();
    expect(second.router.state.location.pathname).toBe('/');
  });
});

describe('SignupPage', () => {
  it('explica o cadastro por e-mail e tem o campo e o botão prontos', async () => {
    anonymousApi();
    renderApp('/signup');
    expect(await screen.findByText(/enviaremos um link para confirmá-lo/)).toBeInTheDocument();
    expect(screen.getByLabelText('E-mail')).toBeEnabled();
    expect(screen.getByRole('button', { name: 'Continuar' })).toBeDisabled();
  });

  it('o Google aparece como "em breve" e desligado (não cria conta por fora do app)', async () => {
    anonymousApi();
    renderApp('/login');
    const google = await screen.findByRole('button', { name: /Continuar com Google/ });
    expect(google).toBeDisabled();
    await waitFor(() => expect(google).toHaveTextContent('em breve'));
  });
});
