import { act, cleanup, screen } from '@testing-library/react';
import { afterEach, beforeAll, describe, expect, it, vi } from 'vitest';
import { setThemeChoice } from './core/theme/useTheme';
import { renderApp, setupTestApp, status, stubApi } from './test-support';

beforeAll(setupTestApp);

afterEach(() => {
  cleanup();
  vi.unstubAllGlobals();
});

describe('App', () => {
  it('mostra texto traduzido, nunca a chave crua, quando as traduções já carregaram', async () => {
    stubApi({ 'GET /version': () => status(500), 'GET /me': () => status(401) });
    renderApp('/');
    expect(await screen.findByRole('heading', { name: 'Entrar' })).toBeInTheDocument();
    expect(screen.queryByText(/^(app|brand|login|auth)\./)).not.toBeInTheDocument();
    expect(screen.getByText(/uma palavra de cada vez/)).toBeInTheDocument();
  });

  it('rota desconhecida mostra "não encontrada" com o caminho de volta', async () => {
    stubApi({ 'GET /version': () => status(500), 'GET /me': () => status(401) });
    renderApp('/nao-existe');
    expect(await screen.findByRole('heading', { name: 'Página não encontrada' })).toBeInTheDocument();
    expect(screen.getByRole('link', { name: 'Voltar ao início' })).toHaveAttribute('href', '/');
  });
});

describe('Tema na raiz', () => {
  it('a escolha escura ou clara troca as classes da raiz e "sistema" as remove', async () => {
    stubApi({ 'GET /version': () => status(500), 'GET /me': () => status(401) });
    const { container } = renderApp('/login');
    await screen.findByRole('heading', { name: 'Entrar' });
    const root = container.firstElementChild as HTMLElement;
    const system = root.className;
    act(() => setThemeChoice('dark'));
    const dark = root.className;
    act(() => setThemeChoice('light'));
    const light = root.className;
    act(() => setThemeChoice('system'));
    expect(new Set([system, dark, light]).size).toBe(3);
    expect(root.className).toBe(system);
  });
});
