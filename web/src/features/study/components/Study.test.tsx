import { cleanup, fireEvent, screen, waitFor } from '@testing-library/react';
import { afterEach, beforeAll, describe, expect, it, vi } from 'vitest';
import { json, me, renderApp, setupTestApp, stubApi, version } from '../../../test-support';
import { accepted, api, house, mother, sentReviews, start } from './studyTestKit';

beforeAll(setupTestApp);

afterEach(() => {
  cleanup();
  vi.unstubAllGlobals();
});

describe('plano', () => {
  it('usa o tempo da meta da semana, sem perguntar, com o fuso do aparelho, e abre a sessão', async () => {
    const calls = api([house]);
    renderApp('/study');
    expect(await screen.findByText('Meta de hoje')).toBeTruthy();
    expect(screen.getByText('15 minutos')).toBeTruthy();
    fireEvent.click(screen.getByRole('button', { name: 'Começar' }));
    expect(await screen.findByRole('heading', { name: 'house' })).toBeTruthy();
    const plan = calls.find((call) => call.path === '/study/plan');
    expect(plan?.headers.get('x-requested-with')).toBe('XMLHttpRequest');
    expect(plan?.body).toMatchObject({ minutes: 15, mode: 'mixed', locale: 'pt-BR' });
  });

  it('sem cartões, avisa que não há o que estudar', async () => {
    api([]);
    await start();
    expect(await screen.findByText('Não há nada para estudar agora. Volte mais tarde.')).toBeTruthy();
  });

  it('se o servidor falhar, avisa e deixa tentar de novo', async () => {
    stubApi({
      'GET /version': () => json(version),
      'GET /me': () => json(me),
      'POST /study/plan': () => new Response(null, { status: 500 }),
    });
    await start();
    expect(await screen.findByText('Não consegui montar a sessão agora. Tente de novo.')).toBeTruthy();
  });
});

describe('sessão completa', () => {
  it('escolher e digitar, dificuldade em um toque, envio em lote e resumo', async () => {
    const calls = api([house, mother]);
    await start();

    fireEvent.click(await screen.findByRole('button', { name: /casa/ }));
    fireEvent.click(screen.getByRole('button', { name: 'Verificar' }));
    expect(await screen.findByText('Correto!')).toBeTruthy();
    fireEvent.click(screen.getByRole('button', { name: 'Médio' }));

    fireEvent.change(await screen.findByLabelText('Sua resposta em inglês'), { target: { value: 'Mother' } });
    fireEvent.click(screen.getByRole('button', { name: 'Verificar' }));
    fireEvent.click(await screen.findByRole('button', { name: /Fácil/ }));

    expect(await screen.findByText('Sessão concluída')).toBeTruthy();
    expect(screen.getByText('+2 palavras conhecidas')).toBeTruthy();

    const reviews = sentReviews(calls);
    expect(reviews).toHaveLength(2);
    expect(reviews[0]).toMatchObject({
      card_id: 'c1',
      session_id: 'sess-1',
      exercise: 'choose',
      chosen_item_id: 'i1',
      client_correct: true,
      difficulty: 'medium',
    });
    expect(reviews[1]).toMatchObject({ card_id: 'c2', exercise: 'type', answer: 'Mother', difficulty: 'easy' });
    expect(new Set(reviews.map((review) => review.id)).size).toBe(2);
  });

  it('errar mostra a resposta, pergunta "Como foi?" e repete o cartão', async () => {
    const calls = api([house]);
    await start();
    fireEvent.click(await screen.findByRole('button', { name: /mãe/ }));
    fireEvent.click(screen.getByRole('button', { name: 'Verificar' }));
    expect(await screen.findByText('Quase!')).toBeTruthy();
    expect(screen.getByText('house', { selector: 'strong' })).toBeTruthy();
    expect(screen.getByText('Como foi?')).toBeTruthy();
    fireEvent.click(screen.getByRole('button', { name: 'Não sabia' }));

    expect(await screen.findByRole('heading', { name: 'house' })).toBeTruthy();
    expect(calls.some((call) => call.path === '/study/reviews')).toBe(false);
  });

  it('dica no digitar limita a nota e só pede para continuar', async () => {
    const calls = api([mother]);
    await start();
    fireEvent.click(await screen.findByRole('button', { name: 'Dica' }));
    expect(screen.getByText('Começa com “m”')).toBeTruthy();
    fireEvent.change(screen.getByLabelText('Sua resposta em inglês'), { target: { value: 'mother' } });
    fireEvent.click(screen.getByRole('button', { name: 'Verificar' }));
    fireEvent.click(await screen.findByRole('button', { name: 'Continuar' }));
    expect(await screen.findByText('Sessão concluída')).toBeTruthy();
    const reviews = sentReviews(calls);
    expect(reviews[0]).toMatchObject({ hint_used: true, difficulty: 'hard' });
  });

  it('"Não sei" conta como erro', async () => {
    api([mother]);
    await start();
    fireEvent.click(await screen.findByRole('button', { name: 'Não sei' }));
    expect(await screen.findByText('Como foi?')).toBeTruthy();
  });
});

describe('teclado', () => {
  it('B escolhe, Enter verifica e 3 responde Fácil', async () => {
    const calls = api([house]);
    await start();
    await screen.findByRole('heading', { name: 'house' });
    fireEvent.keyDown(window, { key: 'b' });
    fireEvent.keyDown(window, { key: 'Enter' });
    expect(await screen.findByText('Correto!')).toBeTruthy();
    fireEvent.keyDown(window, { key: '3' });
    expect(await screen.findByText('Sessão concluída')).toBeTruthy();
    const reviews = sentReviews(calls);
    expect(reviews[0]).toMatchObject({ difficulty: 'easy' });
  });
});

describe('teclado com o foco nos botões', () => {
  it('na pergunta de dificuldade o foco já está no sugerido e 1 responde Difícil', async () => {
    const calls = api([house]);
    await start();
    fireEvent.click(await screen.findByRole('button', { name: /casa/ }));
    fireEvent.click(screen.getByRole('button', { name: 'Verificar' }));
    const suggested = await screen.findByRole('button', { name: /sugerido/ });
    await waitFor(() => expect(document.activeElement).toBe(suggested));
    fireEvent.keyDown(suggested, { key: '1' });
    expect(await screen.findByText('Sessão concluída')).toBeTruthy();
    expect(sentReviews(calls)[0]).toMatchObject({ difficulty: 'hard' });
  });
});

describe('sair e falhar ao salvar', () => {
  it('✕ confirma, envia o que foi respondido e volta ao plano', async () => {
    const calls = api([house, mother]);
    await start();
    fireEvent.click(await screen.findByRole('button', { name: /casa/ }));
    fireEvent.click(screen.getByRole('button', { name: 'Verificar' }));
    fireEvent.click(await screen.findByRole('button', { name: 'Médio' }));
    await screen.findByLabelText('Sua resposta em inglês');

    fireEvent.click(screen.getByRole('button', { name: 'Sair da sessão' }));
    expect(screen.getByText('O que você já respondeu fica salvo.')).toBeTruthy();
    fireEvent.click(screen.getByRole('button', { name: 'Sair' }));

    expect(await screen.findByRole('button', { name: 'Começar' })).toBeTruthy();
    const reviews = sentReviews(calls);
    expect(reviews).toHaveLength(1);
  });

  it('Esc abre e fecha a confirmação; sem respostas, sair não chama o servidor', async () => {
    const calls = api([house]);
    await start();
    await screen.findByRole('heading', { name: 'house' });
    fireEvent.keyDown(window, { key: 'Escape' });
    expect(screen.getByRole('alertdialog')).toBeTruthy();
    fireEvent.keyDown(window, { key: 'Escape' });
    expect(screen.queryByRole('alertdialog')).toBeNull();
    fireEvent.click(screen.getByRole('button', { name: 'Sair da sessão' }));
    fireEvent.click(screen.getByRole('button', { name: 'Sair' }));
    expect(await screen.findByRole('button', { name: 'Começar' })).toBeTruthy();
    expect(calls.some((call) => call.path === '/study/reviews')).toBe(false);
  });

  it('se o envio falhar, a sessão termina com o resumo provisório e a resposta segue na fila', async () => {
    let attempts = 0;
    const calls = api([house], (req) => {
      attempts += 1;
      return attempts === 1 ? new Response(null, { status: 503 }) : accepted(req);
    });
    await start();
    fireEvent.click(await screen.findByRole('button', { name: /casa/ }));
    fireEvent.click(screen.getByRole('button', { name: 'Verificar' }));
    fireEvent.click(await screen.findByRole('button', { name: /Fácil/ }));

    expect(await screen.findByText('Sessão concluída')).toBeTruthy();
    expect(screen.getByText('1 resposta será confirmada quando a conexão voltar.')).toBeTruthy();
    fireEvent.click(screen.getByRole('link', { name: 'Estudar mais' }));

    const bodies = () => calls.filter((call) => call.path === '/study/reviews').map((call) => call.body);
    await waitFor(() => expect(bodies()).toHaveLength(2));
    expect(bodies()[0]).toEqual(bodies()[1]);
  });

  it('sem o pacote na memória (página recarregada), volta ao plano', async () => {
    api([house]);
    renderApp('/study/session/inexistente');
    await waitFor(() => expect(screen.getByRole('button', { name: 'Começar' })).toBeTruthy());
  });
});
