import { onlineManager } from '@tanstack/react-query';
import { cleanup, fireEvent, screen, waitFor } from '@testing-library/react';
import { afterEach, beforeAll, describe, expect, it, vi } from 'vitest';
import { renderApp, setupTestApp } from '../../../test-support';
import { accepted, api, house, mother, start } from './studyTestKit';

beforeAll(setupTestApp);

afterEach(() => {
  cleanup();
  vi.restoreAllMocks();
  onlineManager.setOnline(true);
  vi.unstubAllGlobals();
});

async function answerHouse() {
  fireEvent.click(await screen.findByRole('button', { name: /casa/ }));
  fireEvent.click(screen.getByRole('button', { name: 'Verificar' }));
  fireEvent.click(await screen.findByRole('button', { name: 'Médio' }));
}

async function answerMother() {
  fireEvent.change(await screen.findByLabelText('Sua resposta em inglês'), { target: { value: 'mother' } });
  fireEvent.click(screen.getByRole('button', { name: 'Verificar' }));
  fireEvent.click(await screen.findByRole('button', { name: /Fácil/ }));
}

describe('sem conexão', () => {
  it('a sessão continua, termina com resumo provisório e reenvia ao voltar, com os mesmos UUIDs', async () => {
    let down = true;
    const calls = api([house, mother], (req) => {
      if (down) throw new TypeError('Failed to fetch');
      return accepted(req);
    });
    await start();
    await screen.findByRole('heading', { name: 'house' });
    const onLine = vi.spyOn(navigator, 'onLine', 'get').mockReturnValue(false);
    onlineManager.setOnline(false); // como no navegador: o TanStack pausa mutações sem rede
    fireEvent(window, new Event('offline'));
    expect(await screen.findByText(/Sem conexão/)).toBeTruthy();

    await answerHouse();
    expect(await screen.findByText('1 resposta aguardando envio')).toBeTruthy();
    await answerMother();

    expect(await screen.findByText('Sessão concluída')).toBeTruthy();
    expect(screen.getByText('2 respostas serão confirmadas quando a conexão voltar.')).toBeTruthy();

    down = false;
    onLine.mockReturnValue(true);
    onlineManager.setOnline(true);
    fireEvent.click(screen.getByRole('link', { name: 'Estudar mais' }));
    window.dispatchEvent(new Event('online'));
    await waitFor(() => expect(screen.queryByText(/aguardando envio/)).toBeNull());
    expect(screen.queryByText(/Sem conexão/)).toBeNull();

    const bodies = calls.filter((call) => call.path === '/study/reviews').map((call) => call.body);
    expect(bodies.length).toBeGreaterThanOrEqual(2);
    expect(bodies.at(-1)).toEqual(bodies[0]);
  });

  it('o plano avisa que não há conexão e, ao tentar começar, mostra a falha em vez de travar', async () => {
    api([house]);
    vi.stubGlobal(
      'fetch',
      vi.fn(async () => Promise.reject(new TypeError('Failed to fetch'))),
    ); // sem rede de verdade
    vi.spyOn(navigator, 'onLine', 'get').mockReturnValue(false);
    onlineManager.setOnline(false);
    renderApp('/study');
    expect(await screen.findByText('Sem conexão. Suas respostas ficam salvas e são enviadas depois.')).toBeTruthy();
    fireEvent.click(screen.getByRole('button', { name: 'Começar' }));
    expect(await screen.findByText(/Não consegui montar a sessão/)).toBeTruthy();
  });
});

describe('recarregar a página', () => {
  it('no meio da sessão, retoma de onde parou', async () => {
    api([house, mother]);
    await start();
    await answerHouse();
    await screen.findByLabelText('Sua resposta em inglês');
    await screen.findByText('1 resposta aguardando envio');
    cleanup();

    renderApp('/study/session/sess-1');
    expect(await screen.findByLabelText('Sua resposta em inglês')).toBeTruthy();
    expect(screen.queryByRole('heading', { name: 'house' })).toBeNull();
  });

  it('depois de terminar, voltar para a sessão leva ao plano (não dá para repetir)', async () => {
    api([house]);
    await start();
    await answerHouse();
    expect(await screen.findByText('Sessão concluída')).toBeTruthy();
    cleanup();

    renderApp('/study/session/sess-1');
    expect(await screen.findByRole('button', { name: 'Começar' })).toBeTruthy();
  });
});
