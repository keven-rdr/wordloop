import { cleanup, fireEvent, screen } from '@testing-library/react';
import { afterEach, beforeAll, describe, expect, it, vi } from 'vitest';
import { json, me, renderApp, setupTestApp, status, stubApi, version } from '../../../test-support';

beforeAll(setupTestApp);

afterEach(() => {
  cleanup();
  vi.unstubAllGlobals();
});

function loginApi(answer: () => Response = () => status(204)) {
  let signedIn = false;
  return stubApi({
    'GET /version': () => json(version),
    'GET /me': () => (signedIn ? json(me) : status(401)),
    'POST /auth/password/login': () => {
      const response = answer();
      signedIn = response.ok;
      return response;
    },
  });
}

async function fill(email: string, password: string) {
  fireEvent.change(await screen.findByLabelText('E-mail'), { target: { value: email } });
  fireEvent.change(screen.getByLabelText('Senha'), { target: { value: password } });
}

describe('Login por e-mail e senha', () => {
  it('envia e-mail e senha com o cabeçalho anti-CSRF e mostra o Olá', async () => {
    const calls = loginApi();
    renderApp('/login');
    await fill(' dono@exemplo.com ', 'uma senha bem comprida');
    fireEvent.click(screen.getByRole('button', { name: 'Entrar' }));
    expect(await screen.findByText('Olá, Maria!')).toBeInTheDocument();
    const login = calls.find((c) => c.path === '/auth/password/login');
    expect(login?.body).toEqual({ email: 'dono@exemplo.com', password: 'uma senha bem comprida' });
    expect(login?.headers.get('X-Requested-With')).toBe('XMLHttpRequest');
  });

  it('o botão só habilita com e-mail e senha preenchidos', async () => {
    loginApi();
    renderApp('/login');
    const submit = await screen.findByRole('button', { name: 'Entrar' });
    expect(submit).toBeDisabled();
    await fill('dono@exemplo.com', 'x');
    expect(submit).toBeEnabled();
  });

  it('mostra e oculta a senha', async () => {
    loginApi();
    renderApp('/login');
    const field = await screen.findByLabelText('Senha');
    expect(field).toHaveAttribute('type', 'password');
    fireEvent.click(screen.getByRole('button', { name: 'Mostrar senha' }));
    expect(field).toHaveAttribute('type', 'text');
    fireEvent.click(screen.getByRole('button', { name: 'Ocultar senha' }));
    expect(field).toHaveAttribute('type', 'password');
  });
});

describe('Login por e-mail e senha: recusas', () => {
  it('401 mostra "e-mail ou senha incorretos" sem dizer qual dos dois e mantém o que foi digitado', async () => {
    loginApi(() => status(401));
    renderApp('/login');
    await fill('dono@exemplo.com', 'errada errada errada');
    fireEvent.click(screen.getByRole('button', { name: 'Entrar' }));
    expect(await screen.findByRole('alert')).toHaveTextContent('E-mail ou senha incorretos');
    expect(screen.getByLabelText('E-mail')).toHaveValue('dono@exemplo.com');
  });

  it('403 manda entrar com a passkey e 429 pede para esperar', async () => {
    let code = 403;
    loginApi(() => status(code));
    renderApp('/login');
    await fill('dono@exemplo.com', 'uma senha bem comprida');
    fireEvent.click(screen.getByRole('button', { name: 'Entrar' }));
    expect(await screen.findByRole('alert')).toHaveTextContent('exige 2 fatores');
    code = 429;
    fireEvent.click(screen.getByRole('button', { name: 'Entrar' }));
    expect(await screen.findByText(/Muitas tentativas/)).toBeInTheDocument();
  });

  it('erro de servidor mostra aviso genérico', async () => {
    loginApi(() => status(500));
    renderApp('/login');
    await fill('dono@exemplo.com', 'uma senha bem comprida');
    fireEvent.click(screen.getByRole('button', { name: 'Entrar' }));
    expect(await screen.findByRole('alert')).toHaveTextContent('Não foi possível entrar');
  });

  it('a senha digitada nunca aparece na barra de endereço', async () => {
    loginApi();
    renderApp('/login');
    await fill('dono@exemplo.com', 'uma senha bem comprida');
    fireEvent.click(screen.getByRole('button', { name: 'Entrar' }));
    await screen.findByText('Olá, Maria!');
    expect(window.location.href).not.toContain('senha');
    expect(window.location.search).toBe('');
  });
});
