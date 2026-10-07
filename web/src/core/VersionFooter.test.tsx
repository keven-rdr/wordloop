import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { cleanup, render, screen } from '@testing-library/react';
import { afterEach, beforeAll, describe, expect, it, vi } from 'vitest';
import { configureApi } from './api';
import { setupI18n } from './i18n';
import { VersionFooter } from './VersionFooter';

beforeAll(async () => {
  configureApi('http://localhost');
  await setupI18n('pt-BR');
});

afterEach(() => {
  cleanup();
  vi.unstubAllGlobals();
});

function renderFooter() {
  const client = new QueryClient({ defaultOptions: { queries: { retry: false } } });
  render(
    <QueryClientProvider client={client}>
      <VersionFooter />
    </QueryClientProvider>,
  );
}

describe('VersionFooter', () => {
  it('mostra web, API, ambiente e commit quando a API responde', async () => {
    const json = JSON.stringify({ api: '0.1.0', commit: 'abc1234', env: 'tst' });
    vi.stubGlobal(
      'fetch',
      vi.fn(async () => new Response(json, { headers: { 'content-type': 'application/json' } })),
    );
    renderFooter();
    expect(await screen.findByText(/API v0\.1\.0 · tst/)).toBeInTheDocument();
  });

  it('mostra os dois commits quando web e API vem de commits diferentes', async () => {
    const json = JSON.stringify({ api: '0.1.0', commit: 'fffffff', env: 'tst' });
    vi.stubGlobal(
      'fetch',
      vi.fn(async () => new Response(json, { headers: { 'content-type': 'application/json' } })),
    );
    renderFooter();
    expect(await screen.findByText(/\/fffffff$/)).toBeInTheDocument();
  });

  it('mostra so a versao do web quando a API falha', async () => {
    vi.stubGlobal(
      'fetch',
      vi.fn(async () => new Response('erro', { status: 500 })),
    );
    renderFooter();
    expect(await screen.findByTestId('version-footer')).toHaveTextContent(/^v/);
    expect(screen.getByTestId('version-footer')).not.toHaveTextContent('API');
  });
});
