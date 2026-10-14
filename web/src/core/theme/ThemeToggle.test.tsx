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
