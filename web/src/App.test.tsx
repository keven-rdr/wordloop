import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { cleanup, render, screen } from '@testing-library/react';
import { afterEach, beforeAll, describe, expect, it, vi } from 'vitest';
import { App } from './App';
import { configureApi } from './core/api';
import { setupI18n } from './core/i18n';

beforeAll(async () => {
  configureApi('http://localhost');
  await setupI18n('pt-BR');
});

afterEach(() => {
  cleanup();
  vi.unstubAllGlobals();
});

describe('App', () => {
  it('mostra texto traduzido, nunca a chave crua, quando as traduções já carregaram', () => {
    vi.stubGlobal(
      'fetch',
      vi.fn(async () => new Response('erro', { status: 500 })),
    );
    render(
      <QueryClientProvider client={new QueryClient({ defaultOptions: { queries: { retry: false } } })}>
        <App />
      </QueryClientProvider>,
    );
    expect(screen.getByRole('heading', { name: 'wordloop' })).toBeInTheDocument();
    expect(screen.queryByText(/^app\./)).not.toBeInTheDocument();
    expect(screen.getByText(/vocabulário de inglês/)).toBeInTheDocument();
  });
});
