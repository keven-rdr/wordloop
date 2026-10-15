import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { cleanup, render, screen } from '@testing-library/react';
import { afterEach, beforeAll, describe, expect, it, vi } from 'vitest';
import { configureApi } from './api';
import { setupI18n } from './i18n';
import { VersionFooter } from './VersionFooter';

// O web e compilado com VERSION e COMMIT; aqui cada teste escolhe o seu.
const web = vi.hoisted(() => ({ version: '0.1.0', commit: 'abc1234' }));
vi.mock('./version', () => ({
  get webVersion() {
    return web.version;
  },
  get webCommit() {
    return web.commit;
  },
}));

const FULL = 'abc1234def5678abc1234def5678abc1234def56';

beforeAll(async () => {
  configureApi('http://localhost');
  await setupI18n('pt-BR');
});

afterEach(() => {
  cleanup();
  vi.unstubAllGlobals();
  web.version = '0.1.0';
  web.commit = 'abc1234';
});

function stubVersion(body: object) {
  vi.stubGlobal(
    'fetch',
    vi.fn(async () => new Response(JSON.stringify(body), { headers: { 'content-type': 'application/json' } })),
  );
}

function renderFooter() {
  const client = new QueryClient({ defaultOptions: { queries: { retry: false } } });
  render(
    <QueryClientProvider client={client}>
      <VersionFooter />
    </QueryClientProvider>,
  );
}

describe('VersionFooter', () => {
  it('release: mostra versoes, ambiente e UM commit quando o do web e prefixo do commit completo da API', async () => {
    stubVersion({ api: '0.1.0', commit: FULL, env: 'tst' });
    renderFooter();
    expect(await screen.findByText('v0.1.0 · API v0.1.0 · tst · abc1234')).toBeInTheDocument();
  });

  it('release: mostra os dois commits quando web e API vem de commits diferentes', async () => {
    stubVersion({ api: '0.1.0', commit: 'fffffff0000', env: 'tst' });
    renderFooter();
    expect(await screen.findByText('v0.1.0 · API v0.1.0 · tst · abc1234/fffffff')).toBeInTheDocument();
  });

  it('build comum: mostra o commit curto no lugar da versao, sem repetir o hash', async () => {
    web.version = 'abc1234';
    stubVersion({ api: 'abc1234', commit: FULL, env: 'tst' });
    renderFooter();
    expect(await screen.findByText('web abc1234 · API abc1234 · tst')).toBeInTheDocument();
  });

  it('mostra so o web quando a API falha', async () => {
    vi.stubGlobal(
      'fetch',
      vi.fn(async () => new Response('erro', { status: 500 })),
    );
    renderFooter();
    expect(await screen.findByText('v0.1.0 · abc1234')).toBeInTheDocument();
    expect(screen.getByTestId('version-footer')).not.toHaveTextContent('API');
  });

  it('build comum com a API fora do ar: so o commit do web', async () => {
    web.version = 'abc1234';
    vi.stubGlobal(
      'fetch',
      vi.fn(async () => new Response('erro', { status: 500 })),
    );
    renderFooter();
    expect(await screen.findByText('web abc1234')).toBeInTheDocument();
  });
});
