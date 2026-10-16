import { cleanup, fireEvent, render, screen, waitFor } from '@testing-library/react';
import { afterEach, beforeAll, describe, expect, it, vi } from 'vitest';
import { json, renderApp, setupTestApp, status, stubApi, version } from '../../../test-support';
import { Avatar, initialOf } from './Avatar';

beforeAll(setupTestApp);

afterEach(() => {
  cleanup();
  vi.unstubAllGlobals();
});

const me = {
  id: '00000000-0000-0000-0000-000000000001',
  display_name: 'ana souza',
  role: 'user',
  two_factor_enabled: false,
};

function api(extra: Record<string, () => Response> = {}) {
  return stubApi({
    'GET /version': () => json(version),
    'GET /me': () => json(me),
    'GET /catalog/units': () => json([]),
    'GET /catalog/credits': () => json([]),
    ...extra,
  });
}

describe('Avatar', () => {
  it('sem foto mostra a inicial em maiúscula', () => {
    const { container } = render(<Avatar name="ana souza" />);
    expect(screen.getByText('A')).toBeInTheDocument();
    expect(container.querySelector('img')).toBeNull();
  });

  it('com foto mostra a imagem e, se ela falhar, volta para a inicial', () => {
    const { container } = render(<Avatar name="Bia" pictureUrl="https://lh3.googleusercontent.com/a/x" />);
    const img = container.querySelector('img');
    expect(img).toHaveAttribute('src', 'https://lh3.googleusercontent.com/a/x');
    fireEvent.error(img as Element);
    expect(container.querySelector('img')).toBeNull();
    expect(screen.getByText('B')).toBeInTheDocument();
  });

  it('nome vazio vira interrogação', () => {
    expect(initialOf('   ')).toBe('?');
    expect(initialOf('émerson')).toBe('É');
  });
});

describe('Menu inferior', () => {
  it('mostra Perfil (com a foto), Início e Palavras, e marca o item da tela atual', async () => {
    api({ 'GET /me': () => json({ ...me, picture_url: 'https://lh3.googleusercontent.com/a/x' }) });
    const { container } = renderApp('/words');
    const nav = await screen.findByRole('navigation', { name: 'Menu principal' });
    expect(nav.querySelector('img')).toHaveAttribute('src', 'https://lh3.googleusercontent.com/a/x');
    expect(container.querySelectorAll('nav a')).toHaveLength(3);
    expect(screen.getByRole('link', { name: 'Perfil' })).toHaveAttribute('href', '/settings');
    expect(screen.getByRole('link', { name: 'Início' })).toHaveAttribute('href', '/');
    const words = screen.getByRole('link', { name: 'Palavras' });
    expect(words).toHaveAttribute('aria-current', 'page');
    expect(screen.getByRole('link', { name: 'Início' })).not.toHaveAttribute('aria-current');
  });

  it('em /words/credits o menu continua e dá para voltar às palavras e ao início', async () => {
    api();
    const { router } = renderApp('/words/credits');
    fireEvent.click(await screen.findByRole('link', { name: 'Início' }));
    await waitFor(() => expect(router.state.location.pathname).toBe('/'));
    fireEvent.click(await screen.findByRole('link', { name: 'Palavras' }));
    await waitFor(() => expect(router.state.location.pathname).toBe('/words'));
  });

  it('Perfil abre as configurações e fica marcado também nas passkeys', async () => {
    api({ 'GET /me/passkeys': () => json([]) });
    const { router } = renderApp('/passkeys');
    expect(await screen.findByRole('link', { name: 'Perfil' })).toHaveAttribute('aria-current', 'page');
    fireEvent.click(screen.getByRole('link', { name: 'Perfil' }));
    await waitFor(() => expect(router.state.location.pathname).toBe('/settings'));
  });

  it('sem sessão não há menu', async () => {
    api({ 'GET /me': () => status(401) });
    renderApp('/words');
    expect(await screen.findByRole('heading', { name: 'Entrar' })).toBeInTheDocument();
    expect(screen.queryByRole('navigation', { name: 'Menu principal' })).not.toBeInTheDocument();
  });
});
