import { cleanup, fireEvent, screen, waitFor } from '@testing-library/react';
import { afterEach, beforeAll, beforeEach, describe, expect, it, vi } from 'vitest';
import { json, me, renderApp, setupTestApp, status, stubApi, version } from '../../../test-support';

beforeAll(setupTestApp);

let saved: { name: string; type: string }[] = [];

beforeEach(() => {
  saved = [];
  // O jsdom não tem estas duas funções; define só elas, sem trocar o URL inteiro (o roteador e o fetch precisam do construtor).
  Object.defineProperty(URL, 'createObjectURL', {
    configurable: true,
    value: (blob: Blob) => {
      saved.push({ name: '', type: blob.type });
      return 'blob:fake';
    },
  });
  Object.defineProperty(URL, 'revokeObjectURL', { configurable: true, value: () => undefined });
  vi.spyOn(HTMLAnchorElement.prototype, 'click').mockImplementation(function click(this: HTMLAnchorElement) {
    const last = saved.at(-1);
    if (last) last.name = this.download;
  });
});

afterEach(() => {
  cleanup();
  vi.restoreAllMocks();
  vi.unstubAllGlobals();
  Reflect.deleteProperty(URL, 'createObjectURL');
  Reflect.deleteProperty(URL, 'revokeObjectURL');
});

const data = { schema_version: 1, profile: { display_name: 'Maria' }, reviews: [] };

type Options = { erase?: (req: Request) => Response; exported?: () => Response; signedIn?: boolean };

// Como no servidor de verdade: depois de sair da conta, o /me passa a responder 401.
function api({ erase = () => status(204), exported = () => json(data), signedIn = true }: Options = {}) {
  let active = signedIn;
  return stubApi({
    'GET /version': () => json(version),
    'GET /me': () => (active ? json(me) : status(401)),
    'GET /me/export': exported,
    'POST /me/erase': erase,
    'POST /auth/logout': () => {
      active = false;
      return status(204);
    },
    'GET /auth/providers': () => json({ providers: [] }),
  });
}

const erasing = (calls: ReturnType<typeof api>) => calls.filter((call) => call.path === '/me/erase');

async function openDialog() {
  fireEvent.click(await screen.findByRole('button', { name: 'Excluir minha conta…' }));
  return screen.findByRole('alertdialog');
}

describe('Seus dados: exportar', () => {
  it('mostra o que cada seção faz e avisa que a exclusão não tem volta', async () => {
    api();
    renderApp('/account/data');
    expect(await screen.findByRole('heading', { name: 'Seus dados' })).toBeTruthy();
    expect(screen.getByText(/Não inclui senhas nem chaves/)).toBeTruthy();
    expect(screen.getByText(/Não tem volta/)).toBeTruthy();
  });

  it('baixa o JSON com o nome do dia e avisa quando terminou', async () => {
    const calls = api();
    renderApp('/account/data');
    fireEvent.click(await screen.findByRole('button', { name: 'Baixar meus dados' }));
    expect(await screen.findByText('Pronto! O arquivo foi baixado.')).toBeTruthy();
    expect(calls.some((call) => call.path === '/me/export')).toBe(true);
    expect(saved).toHaveLength(1);
    expect(saved[0]?.type).toBe('application/json');
    expect(saved[0]?.name).toMatch(/^wordloop-dados-\d{4}-\d{2}-\d{2}\.json$/);
  });

  it('se o servidor falhar, avisa e não entrega arquivo', async () => {
    api({ exported: () => status(500) });
    renderApp('/account/data');
    fireEvent.click(await screen.findByRole('button', { name: 'Baixar meus dados' }));
    expect(await screen.findByText('Não consegui baixar agora. Tente de novo.')).toBeTruthy();
    expect(saved).toHaveLength(0);
  });

  it('sem sessão, leva ao login', async () => {
    api({ signedIn: false });
    renderApp('/account/data');
    await waitFor(() => expect(screen.getByText(/Ao continuar, você aceita/)).toBeTruthy());
  });
});

describe('Seus dados: excluir a conta', () => {
  it('o diálogo começa em Cancelar, só libera com o e-mail digitado e Esc fecha sem apagar', async () => {
    const calls = api();
    renderApp('/account/data');
    const dialog = await openDialog();
    await waitFor(() => expect(document.activeElement).toBe(screen.getByRole('button', { name: 'Cancelar' })));
    const confirm = screen.getByRole('button', { name: 'Excluir para sempre' }) as HTMLButtonElement;
    expect(confirm.disabled).toBe(true);
    fireEvent.change(screen.getByLabelText('E-mail da conta'), { target: { value: 'maria@example.com' } });
    expect(confirm.disabled).toBe(false);
    expect(dialog).toBeTruthy();
    fireEvent.keyDown(window, { key: 'Escape' });
    expect(screen.queryByRole('alertdialog')).toBeNull();
    expect(erasing(calls)).toHaveLength(0);
  });

  it('confirmar manda o e-mail digitado, sai da conta e volta ao login', async () => {
    const calls = api();
    renderApp('/account/data');
    await openDialog();
    fireEvent.change(screen.getByLabelText('E-mail da conta'), { target: { value: 'maria@example.com' } });
    fireEvent.click(screen.getByRole('button', { name: 'Excluir para sempre' }));

    await waitFor(() => expect(screen.getByText(/Ao continuar, você aceita/)).toBeTruthy());
    expect(erasing(calls)[0]?.body).toEqual({ confirmation: 'maria@example.com' });
    expect(erasing(calls)[0]?.headers.get('X-Requested-With')).toBe('XMLHttpRequest');
    expect(calls.some((call) => call.path === '/auth/logout')).toBe(true);
  });

  it('Cancelar fecha e não apaga nada', async () => {
    const calls = api();
    renderApp('/account/data');
    await openDialog();
    fireEvent.click(screen.getByRole('button', { name: 'Cancelar' }));
    expect(screen.queryByRole('alertdialog')).toBeNull();
    expect(erasing(calls)).toHaveLength(0);
  });
});
