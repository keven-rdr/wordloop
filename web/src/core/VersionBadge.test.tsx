import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { cleanup, render, screen } from '@testing-library/react';
import { afterEach, beforeAll, describe, expect, it, vi } from 'vitest';
import { configureApi } from './api';
import { setupI18n } from './i18n';
import { VersionBadge } from './VersionBadge';

// O web e compilado com VERSION; aqui cada teste escolhe a sua.
const web = vi.hoisted(() => ({ version: '1.0.162' }));
vi.mock('./version', () => ({
  get webVersion() {
    return web.version;
  },
}));

beforeAll(async () => {
  configureApi('http://localhost');
  await setupI18n('pt-BR');
});

afterEach(() => {
  cleanup();
  vi.unstubAllGlobals();
  web.version = '1.0.162';
});

const ok = (body: object) => new Response(JSON.stringify(body), { headers: { 'content-type': 'application/json' } });

function stubVersion(api: string | undefined) {
  vi.stubGlobal(
    'fetch',
    vi.fn(async () => {
      return api ? ok({ api, commit: 'abc1234', env: 'tst' }) : new Response('erro', { status: 500 });
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
  it('web e API na mesma versao: mostra so v1.0.162, sem hash, API nem ambiente', async () => {
    stubVersion('1.0.162');
    renderBadge();
    await vi.waitFor(() => expect(screen.getByTestId('version-badge')).toHaveTextContent(/^v1\.0\.162$/));
  });

  it('API em outra versao: mostra as duas', async () => {
    stubVersion('1.0.150');
    renderBadge();
    expect(await screen.findByText('v1.0.162 · API v1.0.150')).toBeInTheDocument();
  });

  it('mostra so o web quando a API falha', async () => {
    stubVersion(undefined);
    renderBadge();
    expect(await screen.findByText('v1.0.162')).toBeInTheDocument();
    expect(screen.getByTestId('version-badge')).not.toHaveTextContent('API');
  });

  it('em dev local mostra dev', async () => {
    web.version = 'dev';
    stubVersion('dev');
    renderBadge();
    expect(await screen.findByText('dev')).toBeInTheDocument();
  });
});
