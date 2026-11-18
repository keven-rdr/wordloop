import { cleanup, fireEvent, screen, waitFor } from '@testing-library/react';
import { afterEach, beforeAll, describe, expect, it, vi } from 'vitest';
import { json, me, renderApp, setupTestApp, status, stubApi, version } from '../../../test-support';

beforeAll(setupTestApp);

afterEach(() => {
  cleanup();
  vi.unstubAllGlobals();
});

// Como no servidor de verdade: sair da conta (ou a sessão ter acabado) faz o /me passar a responder 401.
function api(code: number) {
  let signedIn = true;
  return stubApi({
    'GET /version': () => json(version),
    'GET /me': () => (signedIn ? json(me) : status(401)),
    'POST /me/erase': () => {
      if (code === 401) signedIn = false;
      return status(code);
    },
    'POST /auth/logout': () => {
      signedIn = false;
      return status(204);
    },
    'GET /auth/providers': () => json({ providers: [] }),
  });
}

async function tryToErase() {
  renderApp('/account/data');
  fireEvent.click(await screen.findByRole('button', { name: 'Excluir minha conta…' }));
  fireEvent.change(await screen.findByLabelText('E-mail da conta'), { target: { value: 'maria@example.com' } });
  fireEvent.click(screen.getByRole('button', { name: 'Excluir para sempre' }));
}

describe('quando a exclusão não acontece', () => {
  it.each([
    [400, 'O e-mail digitado não é o da conta.'],
    [403, 'Contas de administrador não se apagam por aqui.'],
    [500, 'Não consegui excluir agora. Nada foi apagado. Tente de novo.'],
  ])('status %i: explica e deixa o diálogo aberto', async (code, message) => {
    api(code);
    await tryToErase();
    expect(await screen.findByText(message)).toBeTruthy();
    expect(screen.getByRole('alertdialog')).toBeTruthy();
  });

  it('login antigo (428): pede para entrar de novo e o botão sai da conta', async () => {
    const calls = api(428);
    await tryToErase();
    expect(await screen.findByText('Por segurança, entre de novo e repita o pedido.')).toBeTruthy();
    fireEvent.click(screen.getByRole('button', { name: 'Sair e entrar de novo' }));
    await waitFor(() => expect(screen.getByText(/Ao continuar, você aceita/)).toBeTruthy());
    expect(calls.some((call) => call.path === '/auth/logout')).toBe(true);
  });

  it('sessão acabada (401): leva ao login', async () => {
    api(401);
    await tryToErase();
    await waitFor(() => expect(screen.getByText(/Ao continuar, você aceita/)).toBeTruthy());
  });

  it('depois de um erro dá para corrigir e tentar de novo', async () => {
    const calls = api(400);
    await tryToErase();
    await screen.findByText('O e-mail digitado não é o da conta.');
    fireEvent.change(screen.getByLabelText('E-mail da conta'), { target: { value: 'maria@example.org' } });
    fireEvent.click(screen.getByRole('button', { name: 'Excluir para sempre' }));
    await waitFor(() => expect(calls.filter((call) => call.path === '/me/erase')).toHaveLength(2));
  });
});
