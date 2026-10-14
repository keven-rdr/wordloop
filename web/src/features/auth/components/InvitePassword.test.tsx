import { cleanup, fireEvent, screen } from '@testing-library/react';
import { afterEach, beforeAll, describe, expect, it, vi } from 'vitest';
import { json, me, renderApp, setupTestApp, status, stubApi, version } from '../../../test-support';

beforeAll(setupTestApp);

afterEach(() => {
  cleanup();
  vi.unstubAllGlobals();
  window.history.replaceState(null, '', '/');
});

const strong = 'uma senha bem comprida';

function inviteApi(answer: () => Response = () => status(204)) {
  let signedIn = false;
  return stubApi({
    'GET /version': () => json(version),
    'GET /me': () => (signedIn ? json(me) : status(401)),
    'POST /auth/password/set': () => {
      const response = answer();
      signedIn = response.ok;
      return response;
    },
  });
}

async function openAndFill(password: string, again: string) {
  window.history.replaceState(null, '', '/invite#segredo-do-convite');
  renderApp('/invite');
  fireEvent.change(await screen.findByLabelText('Nova senha'), { target: { value: password } });
  fireEvent.change(screen.getByLabelText('Repita a senha'), { target: { value: again } });
  fireEvent.click(screen.getByRole('button', { name: 'Definir senha e entrar' }));
}

describe('Convite: definir a senha', () => {
  it('manda o convite e a senha, entra e vai para o início', async () => {
    const calls = inviteApi();
    await openAndFill(strong, strong);
    expect(await screen.findByText('Olá, Maria!')).toBeInTheDocument();
    const set = calls.find((c) => c.path === '/auth/password/set');
    expect(set?.body).toEqual({ invite: 'segredo-do-convite', password: strong });
    expect(set?.headers.get('X-Requested-With')).toBe('XMLHttpRequest');
    expect(window.location.href).not.toContain('segredo-do-convite');
  });

  it('senha curta ou diferente da confirmação é barrada sem chamar a API', async () => {
    const calls = inviteApi();
    await openAndFill('curta', 'curta');
    expect(await screen.findByRole('alert')).toHaveTextContent('pelo menos 12 caracteres');
    fireEvent.change(screen.getByLabelText('Nova senha'), { target: { value: strong } });
    fireEvent.change(screen.getByLabelText('Repita a senha'), { target: { value: `${strong}!` } });
    fireEvent.click(screen.getByRole('button', { name: 'Definir senha e entrar' }));
    expect(await screen.findByText('As duas senhas não são iguais.')).toBeInTheDocument();
    expect(calls.some((c) => c.path === '/auth/password/set')).toBe(false);
  });

  it('422 da API mostra a regra da senha e 401 mostra que o convite não vale mais', async () => {
    let code = 422;
    inviteApi(() => status(code));
    await openAndFill(strong, strong);
    expect(await screen.findByRole('alert')).toHaveTextContent('Senha fraca');
    code = 401;
    fireEvent.click(screen.getByRole('button', { name: 'Definir senha e entrar' }));
    expect(await screen.findByText(/O convite não vale mais/)).toBeInTheDocument();
  });

  it('conta com 2FA: a senha é gravada sem sessão e a tela avisa para entrar com a passkey', async () => {
    inviteApi();
    stubApi({
      'GET /version': () => json(version),
      'GET /me': () => status(401),
      'POST /auth/password/set': () => status(204),
    });
    await openAndFill(strong, strong);
    expect(await screen.findByRole('status')).toHaveTextContent('entre com a sua passkey');
    expect(screen.getByRole('link', { name: 'Ir para o login' })).toHaveAttribute('href', '/login');
  });

  it('seis emojis são só 6 caracteres: a tela barra antes de chamar a API', async () => {
    const calls = inviteApi();
    await openAndFill('😀😀😀😀😀😀', '😀😀😀😀😀😀');
    expect(await screen.findByRole('alert')).toHaveTextContent('pelo menos 12 caracteres');
    expect(calls.some((c) => c.path === '/auth/password/set')).toBe(false);
  });

  it('a passkey continua disponível como alternativa', async () => {
    inviteApi();
    window.history.replaceState(null, '', '/invite#segredo');
    renderApp('/invite');
    expect(await screen.findByRole('button', { name: 'Criar minha passkey' })).toBeInTheDocument();
  });
});
