import { act, cleanup, fireEvent, renderHook, screen, waitFor, within } from '@testing-library/react';
import { afterEach, beforeAll, describe, expect, it, vi } from 'vitest';
import { json, renderApp, setupTestApp, stubApi, version } from '../../../test-support';
import { desktopQuery } from '../tokens/breakpoints';
import { useMedia } from '../useMedia';
import { isActive, labelNamespaces } from './items';
import type { NavItem } from './types';

beforeAll(setupTestApp);

afterEach(() => {
  cleanup();
  vi.unstubAllGlobals();
});

const me = { id: '00000000-0000-0000-0000-000000000001', display_name: 'ana', role: 'user', two_factor_enabled: false };

function api() {
  return stubApi({
    'GET /version': () => json(version),
    'GET /me': () => json(me),
    'GET /catalog/units': () => json([]),
    'GET /catalog/credits': () => json([]),
  });
}

// `matchMedia` de mentira: só a consulta do desktop responde, e `resize` avisa quem estiver ouvindo.
function stubMedia(desktop: boolean) {
  const state = { desktop };
  const listeners = new Set<() => void>();
  vi.stubGlobal('matchMedia', (query: string) => ({
    get matches() {
      return query === desktopQuery && state.desktop;
    },
    addEventListener: (_: string, listener: () => void) => listeners.add(listener),
    removeEventListener: (_: string, listener: () => void) => listeners.delete(listener),
  }));
  return (next: boolean) => {
    state.desktop = next;
    for (const listener of listeners) listener();
  };
}

describe('Menu lateral (desktop)', () => {
  it('a partir de bp.desktop mostra só o menu lateral, com a marca e os mesmos itens', async () => {
    stubMedia(true);
    api();
    renderApp('/words');
    const nav = await screen.findByRole('navigation', { name: 'Menu principal' });
    expect(screen.getAllByRole('navigation', { name: 'Menu principal' })).toHaveLength(1);
    expect(within(nav).getByText('wordloop')).toBeInTheDocument();
    const links = within(nav).getAllByRole('link');
    expect(links.map((link) => link.textContent)).toEqual(['APerfil', 'Início', 'Estudar', 'Ranking', 'Palavras']);
    expect(within(nav).getByRole('link', { name: 'Palavras' })).toHaveAttribute('aria-current', 'page');
  });

  it('navega pelo menu lateral', async () => {
    stubMedia(true);
    api();
    const { router } = renderApp('/words');
    fireEvent.click(await screen.findByRole('link', { name: 'Início' }));
    await waitFor(() => expect(router.state.location.pathname).toBe('/'));
  });

  it('ao estreitar a janela troca para o menu inferior (e volta)', async () => {
    const resize = stubMedia(true);
    api();
    renderApp('/words');
    expect(await screen.findByText('wordloop')).toBeInTheDocument();
    act(() => resize(false));
    await waitFor(() => expect(screen.queryByText('wordloop')).not.toBeInTheDocument());
    expect(screen.getByRole('navigation', { name: 'Menu principal' })).toBeInTheDocument();
    act(() => resize(true));
    expect(await screen.findByText('wordloop')).toBeInTheDocument();
  });
});

describe('itens do menu', () => {
  const item = (to: NavItem['to'], also?: string[]): NavItem => ({
    id: 'x',
    label: 'common:nav.home',
    icon: 'avatar',
    to,
    order: 0,
    also,
  });

  it('o Início só fica marcado em /; os outros, na rota e abaixo dela, e nas de `also`', () => {
    expect(isActive(item('/'), '/')).toBe(true);
    expect(isActive(item('/'), '/words')).toBe(false);
    expect(isActive(item('/words'), '/words/credits')).toBe(true);
    expect(isActive(item('/words'), '/wordsmith')).toBe(false);
    expect(isActive(item('/settings', ['/passkeys']), '/passkeys')).toBe(true);
  });

  it('carrega o namespace de cada rótulo (sem prefixo, common)', () => {
    const items = [item('/'), { ...item('/study'), label: 'leaderboard:nav' }, { ...item('/'), label: 'semPrefixo' }];
    expect(labelNamespaces(items)).toEqual(['common', 'leaderboard']);
  });

  it('sem matchMedia, useMedia diz que não (layout mobile)', () => {
    vi.stubGlobal('matchMedia', undefined);
    expect(renderHook(() => useMedia(desktopQuery)).result.current).toBe(false);
  });
});
