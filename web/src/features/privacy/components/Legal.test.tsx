import { cleanup, fireEvent, screen } from '@testing-library/react';
import i18n from 'i18next';
import { afterEach, beforeAll, describe, expect, it, vi } from 'vitest';
import { json, me, renderApp, setupTestApp, status, stubApi, version } from '../../../test-support';

beforeAll(setupTestApp);

afterEach(async () => {
  cleanup();
  await i18n.changeLanguage('pt-BR');
  vi.unstubAllGlobals();
});

function api(signedIn = false) {
  return stubApi({
    'GET /version': () => json(version),
    'GET /me': () => (signedIn ? json(me) : status(401)),
    'GET /auth/providers': () => json({ providers: [] }),
    'GET /me/settings': () => status(401),
  });
}

describe('páginas públicas de privacidade e termos', () => {
  it('a política abre sem entrar na conta, com o aviso de rascunho e todas as seções', async () => {
    api();
    renderApp('/privacy');
    expect(await screen.findByRole('heading', { name: 'Política de privacidade', level: 1 })).toBeTruthy();
    expect(screen.getByRole('note').textContent).toContain('Rascunho em revisão');
    expect(screen.getAllByRole('heading', { level: 2 })).toHaveLength(9);
    for (const heading of ['Quais dados guardamos', 'Por quanto tempo guardamos', 'Seus direitos', 'Segurança']) {
      expect(screen.getByRole('heading', { name: heading })).toBeTruthy();
    }
  });

  it('diz o que é verdade hoje: sem anúncios, sem venda e exclusão imediata', async () => {
    api();
    renderApp('/privacy');
    expect(await screen.findByText(/não mostramos anúncios/)).toBeTruthy();
    expect(screen.getByText(/tudo é apagado na hora/)).toBeTruthy();
    expect(screen.getByText(/nunca a sua biometria/)).toBeTruthy();
  });

  it('os termos abrem sem entrar, com as 8 seções e a licença do conteúdo', async () => {
    api();
    renderApp('/terms');
    expect(await screen.findByRole('heading', { name: 'Termos de uso', level: 1 })).toBeTruthy();
    expect(screen.getAllByRole('heading', { level: 2 })).toHaveLength(8);
    expect(screen.getByText(/CC BY-SA 4.0/)).toBeTruthy();
  });

  it('uma leva à outra e as duas voltam ao início', async () => {
    api();
    renderApp('/privacy');
    fireEvent.click(await screen.findByRole('link', { name: 'Termos de uso' }));
    expect(await screen.findByRole('heading', { name: 'Termos de uso', level: 1 })).toBeTruthy();
    fireEvent.click(screen.getByRole('link', { name: 'Política de privacidade' }));
    expect(await screen.findByRole('heading', { name: 'Política de privacidade', level: 1 })).toBeTruthy();
    expect(screen.getByRole('link', { name: 'Voltar ao início' })).toBeTruthy();
  });

  it('existe em inglês, com as mesmas seções', async () => {
    api();
    await i18n.changeLanguage('en-US');
    renderApp('/privacy');
    expect(await screen.findByRole('heading', { name: 'Privacy policy', level: 1 })).toBeTruthy();
    expect(screen.getAllByRole('heading', { level: 2 })).toHaveLength(9);
  });
});

describe('os links para esses textos', () => {
  it('o login e o cadastro avisam que continuar é aceitar, com os dois links', async () => {
    api();
    renderApp('/login');
    expect(await screen.findByRole('link', { name: 'Termos de uso' })).toBeTruthy();
    expect(screen.getByRole('link', { name: 'Política de privacidade' })).toBeTruthy();
    cleanup();
    renderApp('/signup');
    expect(await screen.findByRole('link', { name: 'Termos de uso' })).toBeTruthy();
  });

  it('as Configurações levam a Seus dados e a Privacidade e termos', async () => {
    api(true);
    renderApp('/settings');
    expect(await screen.findByRole('link', { name: /Seus dados/ })).toBeTruthy();
    fireEvent.click(screen.getByRole('link', { name: /Privacidade e termos/ }));
    expect(await screen.findByRole('heading', { name: 'Política de privacidade', level: 1 })).toBeTruthy();
  });
});
