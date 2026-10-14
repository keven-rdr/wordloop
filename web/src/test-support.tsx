import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { createMemoryHistory } from '@tanstack/react-router';
import { render } from '@testing-library/react';
import { vi } from 'vitest';
import { App } from './App';
import { configureApi } from './core/api';
import { setupI18n } from './core/i18n';
import { createAppRouter } from './core/router';

export async function setupTestApp() {
  configureApi('http://localhost');
  await setupI18n('pt-BR');
}

export function renderApp(path = '/') {
  const router = createAppRouter(createMemoryHistory({ initialEntries: [path] }));
  const client = new QueryClient({ defaultOptions: { queries: { retry: false } } });
  return {
    router,
    ...render(
      <QueryClientProvider client={client}>
        <App router={router} />
      </QueryClientProvider>,
    ),
  };
}

export type Handler = (req: Request) => Response | Promise<Response>;

export type Call = { method: string; path: string; headers: Headers; body: unknown };

export function stubApi(handlers: Record<string, Handler>) {
  const calls: Call[] = [];
  vi.stubGlobal(
    'fetch',
    vi.fn(async (input: Request | string, init?: RequestInit) => {
      const req = input instanceof Request ? input : new Request(input, init);
      const path = new URL(req.url).pathname.replace(/^\/api\/v1/, '');
      const text = await req.clone().text();
      calls.push({ method: req.method, path, headers: req.headers, body: text ? JSON.parse(text) : undefined });
      const handler = handlers[`${req.method} ${path}`];
      return handler ? handler(req) : new Response(null, { status: 404 });
    }),
  );
  return calls;
}

export const json = (body: unknown, status = 200) => Response.json(body, { status });

export const status = (code: number) => new Response(null, { status: code });

export const me = { id: '3f0c4f4e-8d1e-4a3b-9a55-1d2c3b4a5e6f', display_name: 'Maria', role: 'user' };

export const version = { api: '0.1.0', commit: 'abc1234', env: 'tst' };

type Ceremony = { create?: () => Promise<unknown>; get?: () => Promise<unknown> };

export function stubWebAuthn(ceremony: Ceremony = {}) {
  vi.stubGlobal('PublicKeyCredential', {
    parseCreationOptionsFromJSON: (options: unknown) => options,
    parseRequestOptionsFromJSON: (options: unknown) => options,
  });
  const issued = async () => ({ toJSON: () => ({ id: 'cred', type: 'public-key' }) });
  const create = vi.fn(ceremony.create ?? issued);
  const get = vi.fn(ceremony.get ?? issued);
  const credentials = { create, get };
  Object.defineProperty(navigator, 'credentials', { value: credentials, configurable: true });
  return credentials;
}

export function withoutWebAuthn() {
  vi.stubGlobal('PublicKeyCredential', undefined);
  Object.defineProperty(navigator, 'credentials', { value: undefined, configurable: true });
}
