import { cleanup, fireEvent, render, screen } from '@testing-library/react';
import { afterEach, beforeAll, describe, expect, it, vi } from 'vitest';
import { setupTestApp } from '../../test-support';
import { ThemeToggle } from './ThemeToggle';
import { setThemeChoice, useThemeChoice } from './useTheme';

beforeAll(setupTestApp);

afterEach(() => {
  cleanup();
  setThemeChoice('system');
  window.localStorage.clear();
  vi.unstubAllGlobals();
});

function Choice() {
  return <output data-testid="choice">{useThemeChoice()}</output>;
}

function renderToggle() {
  return render(
    <>
      <ThemeToggle />
      <Choice />
    </>,
  );
}

describe('ThemeToggle', () => {
  it('começa no tema do sistema e alterna claro/escuro, guardando a escolha', async () => {
    vi.stubGlobal('matchMedia', () => ({ matches: false }));
    renderToggle();
    expect(screen.getByTestId('choice')).toHaveTextContent('system');
    fireEvent.click(await screen.findByRole('button', { name: 'Usar tema escuro' }));
    expect(screen.getByTestId('choice')).toHaveTextContent('dark');
    expect(window.localStorage.getItem('wl-theme')).toBe('dark');
    fireEvent.click(screen.getByRole('button', { name: 'Usar tema claro' }));
    expect(screen.getByTestId('choice')).toHaveTextContent('light');
  });

  it('com o sistema escuro oferece o tema claro primeiro', async () => {
    vi.stubGlobal('matchMedia', () => ({ matches: true }));
    renderToggle();
    expect(await screen.findByRole('button', { name: 'Usar tema claro' })).toBeInTheDocument();
  });

  it('sem armazenamento disponível a escolha vale na sessão, sem quebrar', async () => {
    vi.spyOn(Storage.prototype, 'setItem').mockImplementation(() => {
      throw new Error('bloqueado');
    });
    renderToggle();
    fireEvent.click(await screen.findByRole('button', { name: /Usar tema/ }));
    expect(screen.getByTestId('choice')).not.toHaveTextContent('system');
    vi.restoreAllMocks();
  });
});

describe('useTheme: escolha guardada', () => {
  it('lê a escolha guardada ao abrir e ignora valores estranhos', async () => {
    window.localStorage.setItem('wl-theme', 'dark');
    vi.resetModules();
    const fresh = await import('./useTheme');
    const { renderHook } = await import('@testing-library/react');
    expect(renderHook(() => fresh.useThemeChoice()).result.current).toBe('dark');
    window.localStorage.setItem('wl-theme', 'azul');
    vi.resetModules();
    const again = await import('./useTheme');
    expect(renderHook(() => again.useThemeChoice()).result.current).toBe('system');
  });

  it('se ler o armazenamento falhar, usa o tema do sistema', async () => {
    vi.spyOn(Storage.prototype, 'getItem').mockImplementation(() => {
      throw new Error('bloqueado');
    });
    vi.resetModules();
    const fresh = await import('./useTheme');
    const { renderHook } = await import('@testing-library/react');
    expect(renderHook(() => fresh.useThemeChoice()).result.current).toBe('system');
    vi.restoreAllMocks();
  });
});
