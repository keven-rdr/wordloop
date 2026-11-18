import { cleanup, fireEvent, screen, waitFor } from '@testing-library/react';
import i18n from 'i18next';
import { afterEach, beforeAll, beforeEach, describe, expect, it, vi } from 'vitest';
import { setThemeChoice } from '../../../core/theme/useTheme';
import { json, me, renderApp, setupTestApp, status, stubApi, version } from '../../../test-support';

beforeAll(setupTestApp);

// O aparelho dos testes fala português: "usar o idioma do aparelho" deixa a tela em português.
beforeEach(() => {
  vi.spyOn(navigator, 'languages', 'get').mockReturnValue(['pt-BR']);
});

afterEach(async () => {
  cleanup();
  vi.restoreAllMocks();
  setThemeChoice('system');
  window.localStorage.clear();
  await i18n.changeLanguage('pt-BR');
  vi.unstubAllGlobals();
});

const device = Intl.DateTimeFormat().resolvedOptions().timeZone;
const defaults = { ui_locale: 'system', theme: 'system', tz: 'America/Sao_Paulo', tz_auto: true };

type Options = { saved?: Record<string, unknown>; save?: (req: Request) => Response | Promise<Response> };

function api({ saved = defaults, save }: Options = {}) {
  const echo = async (req: Request) => json(await req.clone().json());
  return stubApi({
    'GET /version': () => json(version),
    'GET /me': () => json(me),
    'GET /me/settings': () => (saved === null ? status(500) : json(saved)),
    'PUT /me/settings': save ?? echo,
  });
}

const puts = (calls: ReturnType<typeof api>) =>
  calls.filter((call) => call.method === 'PUT' && call.path === '/me/settings');

describe('tela de preferências', () => {
  it('mostra o que está salvo: idioma e tema do aparelho e o fuso acompanhando o aparelho', async () => {
    api();
    renderApp('/preferences');
    expect(await screen.findByRole('heading', { name: 'Preferências' })).toBeTruthy();
    expect((screen.getByRole('radio', { name: 'Usar o idioma do aparelho' }) as HTMLInputElement).checked).toBe(true);
    expect((screen.getByRole('radio', { name: 'Usar o tema do aparelho' }) as HTMLInputElement).checked).toBe(true);
    expect((screen.getByRole('checkbox', { name: /Seguir o fuso do aparelho/ }) as HTMLInputElement).checked).toBe(
      true,
    );
    expect(screen.getByText(`Detectado: ${device}`)).toBeTruthy();
    expect(screen.queryByLabelText('Escolha o fuso')).toBeNull();
  });

  it('trocar o idioma salva as preferências inteiras e já traduz a tela', async () => {
    const calls = api();
    renderApp('/preferences');
    fireEvent.click(await screen.findByRole('radio', { name: 'English (US)' }));

    await waitFor(() => expect(puts(calls)).toHaveLength(1));
    expect(puts(calls)[0]?.body).toEqual({ ...defaults, ui_locale: 'en-US' });
    expect(puts(calls)[0]?.headers.get('X-Requested-With')).toBe('XMLHttpRequest');
    expect(await screen.findByRole('heading', { name: 'Preferences' })).toBeTruthy();
    expect(screen.getByText('Saved')).toBeTruthy();
  });

  it('trocar o tema salva e aplica na hora', async () => {
    const calls = api();
    renderApp('/preferences');
    fireEvent.click(await screen.findByRole('radio', { name: 'Escuro' }));
    await waitFor(() => expect(puts(calls)).toHaveLength(1));
    expect(puts(calls)[0]?.body).toEqual({ ...defaults, theme: 'dark' });
    await waitFor(() => expect(window.localStorage.getItem('wl-theme')).toBe('dark'));
    expect(await screen.findByText('Salvo')).toBeTruthy();
  });
});

describe('fuso, falhas e acesso à tela de preferências', () => {
  it('desligar "seguir o aparelho" mostra a lista de fusos e escolher um o deixa fixo', async () => {
    const calls = api();
    renderApp('/preferences');
    fireEvent.click(await screen.findByRole('checkbox', { name: /Seguir o fuso do aparelho/ }));
    await waitFor(() => expect(puts(calls)).toHaveLength(1));
    expect(puts(calls)[0]?.body).toEqual({ ...defaults, tz_auto: false });

    const select = (await screen.findByLabelText('Escolha o fuso')) as HTMLSelectElement;
    expect(select.value).toBe('America/Sao_Paulo');
    fireEvent.change(select, { target: { value: 'Asia/Tokyo' } });
    await waitFor(() => expect(puts(calls)).toHaveLength(2));
    expect(puts(calls)[1]?.body).toEqual({ ...defaults, tz: 'Asia/Tokyo', tz_auto: false });
    expect(await screen.findByText(/vira às 4h no fuso Asia\/Tokyo/)).toBeTruthy();
  });

  it('ligar de novo "seguir o aparelho" manda o fuso do aparelho', async () => {
    const calls = api({ saved: { ...defaults, tz: 'Asia/Tokyo', tz_auto: false } });
    renderApp('/preferences');
    fireEvent.click(await screen.findByRole('checkbox', { name: /Seguir o fuso do aparelho/ }));
    await waitFor(() => expect(puts(calls)).toHaveLength(1));
    expect(puts(calls)[0]?.body).toEqual({ ...defaults, tz: device, tz_auto: true });
  });

  it('se salvar falhar, avisa', async () => {
    api({ save: () => status(500) });
    renderApp('/preferences');
    fireEvent.click(await screen.findByRole('radio', { name: 'Escuro' }));
    expect(await screen.findByText('Não consegui salvar agora. Tente de novo.')).toBeTruthy();
  });

  it('se não carregar, avisa', async () => {
    api({ saved: null as unknown as Record<string, unknown> });
    renderApp('/preferences');
    expect(await screen.findByText('Não consegui carregar suas preferências agora.')).toBeTruthy();
  });

  it('as Configurações levam às preferências', async () => {
    api();
    renderApp('/settings');
    fireEvent.click(await screen.findByRole('link', { name: /Preferências/ }));
    expect(await screen.findByRole('heading', { name: 'Preferências' })).toBeTruthy();
  });
});

describe('preferências salvas valem em qualquer tela', () => {
  it('quem entra recebe de volta o idioma e o tema que escolheu', async () => {
    api({ saved: { ...defaults, ui_locale: 'en-US', theme: 'dark' } });
    renderApp('/');
    expect(await screen.findByText('Study now')).toBeTruthy();
    expect(window.localStorage.getItem('wl-theme')).toBe('dark');
  });

  it('sem preferências salvas (idioma "sistema"), o idioma do aparelho segue valendo', async () => {
    api();
    renderApp('/');
    expect(await screen.findByText('Estudar agora')).toBeTruthy();
  });

  it('se o servidor falhar, a tela segue como estava', async () => {
    api({ saved: null as unknown as Record<string, unknown> });
    renderApp('/');
    expect(await screen.findByText('Estudar agora')).toBeTruthy();
  });
});
