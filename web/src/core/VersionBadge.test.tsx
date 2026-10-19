import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { cleanup, render, screen } from '@testing-library/react';
import { afterEach, beforeAll, describe, expect, it, vi } from 'vitest';
import { configureApi } from './api';
import { setupI18n } from './i18n';
import { VersionBadge } from './VersionBadge';

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

function stubVersion(body: object, signedIn = false) {
  vi.stubGlobal(
    'fetch',
    vi.fn(async (input: Request | string) => {
      const url = input instanceof Request ? input.url : input;
      if (url.endsWith('/me')) {
        return signedIn
          ? new Response(JSON.stringify({ id: 'x', display_name: 'Ana', role: 'user', two_factor_enabled: false }), {
              headers: { 'content-type': 'application/json' },
            })
          : new Response(null, { status: 401 });
      }
      return new Response(JSON.stringify(body), { headers: { 'content-type': 'application/json' } });
    }),
  );
}

function renderBadge() {
  const client = new QueryClient({ defaultOptions: { queries: { retry: false } } });
  render(
    <QueryClientProvider client={client}>
      <VersionBadge />
    </QueryClientProvider>,
  );
}

describe('VersionBadge', () => {
  it('release: mostra versoes, ambiente e UM commit quando o do web e prefixo do commit completo da API', async () => {
    stubVersion({ api: '0.1.0', commit: FULL, env: 'tst' });
    renderBadge();
    expect(await screen.findByText('v0.1.0 · API v0.1.0 · tst · abc1234')).toBeInTheDocument();
  });

  it('release: mostra os dois commits quando web e API vem de commits diferentes', async () => {
    stubVersion({ api: '0.1.0', commit: 'fffffff0000', env: 'tst' });
    renderBadge();
    expect(await screen.findByText('v0.1.0 · API v0.1.0 · tst · abc1234/fffffff')).toBeInTheDocument();
  });

  it('sem release: mostra v0.0.0 e o commit', async () => {
    web.version = 'abc1234';
    stubVersion({ api: 'abc1234', commit: FULL, env: 'tst' });
    renderBadge();
    expect(await screen.findByText('v0.0.0 · API v0.0.0 · tst · abc1234')).toBeInTheDocument();
  });

  it('mostra so o web quando a API falha', async () => {
    vi.stubGlobal(
      'fetch',
      vi.fn(async () => new Response('erro', { status: 500 })),
    );
    renderBadge();
    expect(await screen.findByText('v0.1.0 · abc1234')).toBeInTheDocument();
    expect(screen.getByTestId('version-badge')).not.toHaveTextContent('API');
  });

  it('sem release e com a API fora do ar: so o web', async () => {
    web.version = 'abc1234';
    vi.stubGlobal(
      'fetch',
      vi.fn(async () => new Response('erro', { status: 500 })),
    );
    renderBadge();
    expect(await screen.findByText('v0.0.0 · abc1234')).toBeInTheDocument();
  });

  it('com sessao o selo sobe acima do menu e sem sessao fica no canto', async () => {
    stubVersion({ api: '0.1.0', commit: FULL, env: 'tst' }, false);
    renderBadge();
    await screen.findByText(/API v0.1.0/);
    const corner = screen.getByTestId('version-badge').className;
    cleanup();

    stubVersion({ api: '0.1.0', commit: FULL, env: 'tst' }, true);
    renderBadge();
    await screen.findByText(/API v0.1.0/);
    await vi.waitFor(() => expect(screen.getByTestId('version-badge').className).not.toBe(corner));
  });
});
