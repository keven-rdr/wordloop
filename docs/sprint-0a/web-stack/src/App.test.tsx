import { cleanup, fireEvent, render, screen } from '@testing-library/react';
import { afterEach, describe, expect, it } from 'vitest';
import { App } from './App';

afterEach(cleanup);

describe('App (spike)', () => {
  it('renderiza o botão e avança a barra de progresso', () => {
    render(<App />);
    const bar = screen.getByRole('progressbar');
    expect(bar.getAttribute('aria-valuenow')).toBe('20');
    fireEvent.click(screen.getByRole('button', { name: 'Continuar' }));
    expect(screen.getByRole('progressbar').getAttribute('aria-valuenow')).toBe('40');
  });

  it('aplica classes atômicas do StyleX e troca de tema', () => {
    render(<App />);
    const page = screen.getByTestId('page');
    const before = page.className;
    expect(before.length).toBeGreaterThan(0);
    fireEvent.click(screen.getByRole('switch', { name: 'Tema escuro' }));
    expect(screen.getByTestId('page').className).not.toBe(before);
  });
});
