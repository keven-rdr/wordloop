import { cleanup, screen } from '@testing-library/react';
import { afterEach, beforeAll, describe, expect, it, vi } from 'vitest';
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
    expect(await screen.findByRole('heading', { name: 'wordloop' })).toBeInTheDocument();
    expect(screen.queryByText(/^app\./)).not.toBeInTheDocument();
    expect(screen.getByText(/vocabulário de inglês/)).toBeInTheDocument();
  });

  it('rota desconhecida mostra "não encontrada" com o caminho de volta', async () => {
    stubApi({ 'GET /version': () => status(500), 'GET /me': () => status(401) });
    renderApp('/nao-existe');
    expect(await screen.findByRole('heading', { name: 'Página não encontrada' })).toBeInTheDocument();
    expect(screen.getByRole('link', { name: 'Voltar ao início' })).toHaveAttribute('href', '/');
  });
});
