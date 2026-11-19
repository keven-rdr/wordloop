import { cleanup, fireEvent, screen, waitFor, within } from '@testing-library/react';
import { afterEach, beforeAll, describe, expect, it, vi } from 'vitest';
import { setupTestApp } from '../../../test-support';
import { api, house, sentReviews, start } from './studyTestKit';

beforeAll(setupTestApp);

afterEach(() => {
  cleanup();
  vi.unstubAllGlobals();
});

// O servidor manda as peças já embaralhadas, sem pontuação; a pergunta é a tradução.
const sentence = {
  ...house,
  id: 'c5',
  item_id: 'i5',
  direction: 'productive',
  state: 'review',
  exercise: 'assemble',
  term: 'I like the red house.',
  gloss: 'Eu gosto da casa vermelha.',
  tiles: ['red', 'I', 'house', 'the', 'like'],
  options: undefined,
};

const bank = () => within(screen.getByRole('list', { name: 'Palavras para montar a frase' }));
const tap = (word: string) => fireEvent.click(bank().getByRole('button', { name: word }));
const build = (words: string[]) => {
  for (const word of words) tap(word);
};

describe('montar a frase', () => {
  it('pergunta pela tradução, mostra as peças e só libera "Verificar" com a frase completa', async () => {
    api([sentence]);
    await start();
    expect(await screen.findByRole('heading', { name: 'Eu gosto da casa vermelha.' })).toBeTruthy();
    expect(screen.getByText('Monte a frase em inglês')).toBeTruthy();
    expect(screen.getByText('Toque nas palavras abaixo, na ordem certa.')).toBeTruthy();
    expect(screen.queryByText('I like the red house.')).toBeNull();
    expect(bank().getAllByRole('button')).toHaveLength(5);

    const verify = screen.getByRole('button', { name: 'Verificar' }) as HTMLButtonElement;
    build(['I', 'like', 'the', 'red']);
    expect(verify.disabled).toBe(true);
    tap('house');
    expect(verify.disabled).toBe(false);
    expect(bank().queryAllByRole('button')).toHaveLength(0);
  });

  it('acertar: manda as peças juntas, na ordem colocada, e o servidor confirma', async () => {
    const calls = api([sentence]);
    await start();
    await screen.findByRole('heading', { name: 'Eu gosto da casa vermelha.' });
    build(['I', 'like', 'the', 'red', 'house']);
    fireEvent.click(screen.getByRole('button', { name: 'Verificar' }));
    expect(await screen.findByText('Correto!')).toBeTruthy();
    fireEvent.click(screen.getByRole('button', { name: /Fácil/ }));

    expect(await screen.findByText('Sessão concluída')).toBeTruthy();
    expect(sentReviews(calls)[0]).toMatchObject({
      card_id: 'c5',
      exercise: 'assemble',
      answer: 'I like the red house',
      client_correct: true,
    });
  });

  it('errar mostra a frase certa e pergunta "Como foi?"', async () => {
    const calls = api([sentence]);
    await start();
    await screen.findByRole('heading', { name: 'Eu gosto da casa vermelha.' });
    build(['like', 'I', 'the', 'red', 'house']);
    fireEvent.click(screen.getByRole('button', { name: 'Verificar' }));
    expect(await screen.findByText('Como foi?')).toBeTruthy();
    expect(screen.getByText('I like the red house.', { selector: 'strong' })).toBeTruthy();
    fireEvent.click(screen.getByRole('button', { name: /Foi distração/ }));
    // o cartão errado volta para a fila: a pergunta recomeça com o banco cheio, e acertar encerra a sessão
    await waitFor(() => expect(bank().queryAllByRole('button')).toHaveLength(5));
    build(['I', 'like', 'the', 'red', 'house']);
    fireEvent.click(screen.getByRole('button', { name: 'Verificar' }));
    fireEvent.click(await screen.findByRole('button', { name: /Médio/ }));
    expect(await screen.findByText('Sessão concluída')).toBeTruthy();
    expect(sentReviews(calls)[0]).toMatchObject({
      exercise: 'assemble',
      answer: 'like I the red house',
      client_correct: false,
    });
    expect(sentReviews(calls)[1]).toMatchObject({ answer: 'I like the red house', client_correct: true });
  });
});

describe('montar a frase: desfazer', () => {
  it('tocar numa peça colocada a devolve ao banco; "Limpar tudo" devolve todas', async () => {
    api([sentence]);
    await start();
    await screen.findByRole('heading', { name: 'Eu gosto da casa vermelha.' });
    build(['I', 'like']);
    expect(bank().getAllByRole('button')).toHaveLength(3);
    fireEvent.click(screen.getByRole('button', { name: 'Tirar I' }));
    expect(bank().getAllByRole('button')).toHaveLength(4);
    expect(bank().getByRole('button', { name: 'I' })).toBeTruthy();
    fireEvent.click(screen.getByRole('button', { name: 'Limpar tudo' }));
    expect(bank().getAllByRole('button')).toHaveLength(5);
    expect(screen.queryByRole('button', { name: 'Limpar tudo' })).toBeNull();
  });
});

describe('montar a frase: teclado e casos de borda', () => {
  it('teclado: 1 a 9 colocam a peça daquela posição, Backspace desfaz e Enter verifica', async () => {
    const calls = api([sentence]);
    await start();
    await screen.findByRole('heading', { name: 'Eu gosto da casa vermelha.' });
    // banco: red, I, house, the, like → 2 = I
    fireEvent.keyDown(window, { key: '2' });
    expect(fireEvent.keyDown(window, { key: '9' })).toBe(true); // não existe: ignorada, o navegador segue normal
    expect(bank().getAllByRole('button')).toHaveLength(4);
    fireEvent.keyDown(window, { key: 'Backspace' });
    expect(bank().getAllByRole('button')).toHaveLength(5);
    fireEvent.keyDown(window, { key: 'Enter' }); // incompleta: nada acontece
    expect(screen.queryByText('Correto!')).toBeNull();
    // I, like, the, red, house: no banco que vai encolhendo, as posições são 2, 4(like), 3(the)...
    fireEvent.keyDown(window, { key: '2' }); // I      → banco: red, house, the, like
    fireEvent.keyDown(window, { key: '4' }); // like   → banco: red, house, the
    fireEvent.keyDown(window, { key: '3' }); // the    → banco: red, house
    fireEvent.keyDown(window, { key: '1' }); // red    → banco: house
    fireEvent.keyDown(window, { key: '1' }); // house
    // tecla tratada: o navegador não pode ativar o botão que ganha o foco logo depois (a dificuldade sugerida)
    expect(fireEvent.keyDown(window, { key: 'Enter' })).toBe(false);
    expect(await screen.findByText('Correto!')).toBeTruthy();
    fireEvent.click(screen.getByRole('button', { name: /Médio/ }));
    expect(await screen.findByText('Sessão concluída')).toBeTruthy();
    expect(sentReviews(calls)[0]).toMatchObject({ answer: 'I like the red house' });
  });

  it('"Não sei" conta como erro e peças iguais são independentes', async () => {
    const twins = { ...sentence, term: 'No no yes.', gloss: 'Não, não, sim.', tiles: ['no', 'yes', 'no'] };
    const calls = api([twins]);
    await start();
    await screen.findByRole('heading', { name: 'Não, não, sim.' });
    fireEvent.click(bank().getAllByRole('button', { name: 'no' })[0] as HTMLElement);
    expect(bank().getAllByRole('button', { name: 'no' })).toHaveLength(1);
    fireEvent.click(screen.getByRole('button', { name: 'Não sei' }));
    expect(await screen.findByText('Como foi?')).toBeTruthy();
    fireEvent.click(screen.getByRole('button', { name: /Não sabia/ }));
    // volta para a fila; montar certo (no, yes, no? a ordem certa é "No no yes") encerra a sessão
    await waitFor(() => expect(bank().queryAllByRole('button')).toHaveLength(3));
    fireEvent.click(bank().getAllByRole('button', { name: 'no' })[0] as HTMLElement);
    fireEvent.click(bank().getAllByRole('button', { name: 'no' })[0] as HTMLElement);
    tap('yes');
    fireEvent.click(screen.getByRole('button', { name: 'Verificar' }));
    fireEvent.click(await screen.findByRole('button', { name: /Médio/ }));
    expect(await screen.findByText('Sessão concluída')).toBeTruthy();
    expect(sentReviews(calls)[0]).toMatchObject({ exercise: 'assemble', answer: '', client_correct: false });
    expect(sentReviews(calls)[1]).toMatchObject({ answer: 'no no yes', client_correct: true });
  });
});
