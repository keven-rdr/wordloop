import { cleanup, fireEvent, screen } from '@testing-library/react';
import { afterEach, beforeAll, describe, expect, it, vi } from 'vitest';
import { json, me, renderApp, setupTestApp, status, stubApi, version } from '../../../test-support';

beforeAll(setupTestApp);

afterEach(() => {
  cleanup();
  vi.unstubAllGlobals();
  window.history.replaceState(null, '', '/');
});

function emailApi(path: string, answer: () => Response = () => status(202)) {
  return stubApi({
    'GET /version': () => json(version),
    'GET /me': () => status(401),
    [`POST ${path}`]: answer,
  });
}

async function send(address: string) {
  fireEvent.change(await screen.findByLabelText('E-mail'), { target: { value: address } });
  fireEvent.click(screen.getByRole('button', { name: /^(Continuar|Enviar o link)$/ }));
}

describe('Cadastro por e-mail', () => {
  it('manda o e-mail e o idioma com o cabeçalho anti-CSRF e mostra para onde o link foi', async () => {
    const calls = emailApi('/auth/register');
    renderApp('/signup');
    await send('  ana@exemplo.com ');
    expect(await screen.findByRole('status')).toHaveTextContent('Enviamos um link para ana@exemplo.com');
    const call = calls.find((c) => c.path === '/auth/register');
    expect(call?.body).toEqual({ email: 'ana@exemplo.com', locale: 'pt-BR' });
    expect(call?.headers.get('X-Requested-With')).toBe('XMLHttpRequest');
    expect(screen.queryByLabelText('E-mail')).not.toBeInTheDocument();
  });

  it('o botão só habilita com e-mail preenchido', async () => {
    emailApi('/auth/register');
    renderApp('/signup');
    expect(await screen.findByRole('button', { name: 'Continuar' })).toBeDisabled();
  });

  it('cada recusa da API tem a sua mensagem', async () => {
    const cases: Array<[number, string]> = [
      [400, 'Informe um e-mail válido'],
      [429, 'Muitas tentativas'],
      [503, 'não está disponível neste ambiente'],
      [500, 'Não foi possível enviar agora'],
    ];
    for (const [code, message] of cases) {
      cleanup();
      emailApi('/auth/register', () => status(code));
      renderApp('/signup');
      await send('ana@exemplo.com');
      expect(await screen.findByRole('alert')).toHaveTextContent(message);
    }
  });
});

describe('Esqueci a senha', () => {
  it('o login leva à tela e ela pede o link sem dizer se o e-mail existe', async () => {
    const calls = emailApi('/auth/password/forgot');
    renderApp('/login');
    fireEvent.click(await screen.findByRole('link', { name: 'Esqueceu a senha?' }));
    expect(await screen.findByRole('heading', { name: 'Esqueci a senha' })).toBeInTheDocument();
    await send('ninguem@exemplo.com');
    expect(await screen.findByRole('status')).toHaveTextContent(
      'Se existir uma conta com o e-mail ninguem@exemplo.com',
    );
    expect(calls.find((c) => c.path === '/auth/password/forgot')?.body).toEqual({
      email: 'ninguem@exemplo.com',
      locale: 'pt-BR',
    });
  });

  it('quem já está logado é levado para o início e o link de voltar funciona', async () => {
    stubApi({ 'GET /version': () => json(version), 'GET /me': () => json(me) });
    renderApp('/forgot');
    expect(await screen.findByText('Olá, Maria!')).toBeInTheDocument();
    cleanup();
    emailApi('/auth/password/forgot');
    renderApp('/forgot');
    fireEvent.click(await screen.findByRole('link', { name: 'Voltar ao login' }));
    expect(await screen.findByRole('heading', { name: 'Entrar' })).toBeInTheDocument();
  });
});

describe('Links do e-mail (confirmar e redefinir)', () => {
  const strong = 'uma senha bem comprida';

  async function openLink(path: string) {
    window.history.replaceState(null, '', `${path}#segredo-do-link`);
    renderApp(path);
    fireEvent.change(await screen.findByLabelText('Nova senha'), { target: { value: strong } });
    fireEvent.change(screen.getByLabelText('Repita a senha'), { target: { value: strong } });
    fireEvent.click(screen.getByRole('button', { name: 'Definir senha e entrar' }));
  }

  it.each([
    ['/verify', 'Confirmar o e-mail'],
    ['/reset', 'Redefinir a senha'],
  ])('%s esconde o segredo, manda o token e a senha e entra', async (path, title) => {
    let signedIn = false;
    const calls = stubApi({
      'GET /version': () => json(version),
      'GET /me': () => (signedIn ? json(me) : status(401)),
      'POST /auth/password/set': () => {
        signedIn = true;
        return status(204);
      },
    });
    window.history.replaceState(null, '', `${path}#segredo-do-link`);
    renderApp(path);
    expect(await screen.findByRole('heading', { name: title })).toBeInTheDocument();
    fireEvent.change(screen.getByLabelText('Nova senha'), { target: { value: strong } });
    fireEvent.change(screen.getByLabelText('Repita a senha'), { target: { value: strong } });
    fireEvent.click(screen.getByRole('button', { name: 'Definir senha e entrar' }));
    expect(await screen.findByText('Olá, Maria!')).toBeInTheDocument();
    expect(calls.find((c) => c.path === '/auth/password/set')?.body).toEqual({
      token: 'segredo-do-link',
      password: strong,
    });
    expect(window.location.href).not.toContain('segredo-do-link');
  });

  it.each(['/verify', '/reset'])(
    '%s sem segredo no link mostra que o link não vale e não chama a API',
    async (path) => {
      const calls = stubApi({ 'GET /version': () => json(version), 'GET /me': () => status(401) });
      renderApp(path);
      expect(await screen.findByRole('alert')).toHaveTextContent('Este link não é válido');
      expect(calls.some((c) => c.path === '/auth/password/set')).toBe(false);
    },
  );

  it('token vencido (401) mostra que o link não vale mais', async () => {
    stubApi({
      'GET /version': () => json(version),
      'GET /me': () => status(401),
      'POST /auth/password/set': () => status(401),
    });
    await openLink('/reset');
    expect(await screen.findByRole('alert')).toHaveTextContent('O link não vale mais');
  });
});
