import { cleanup, fireEvent, screen } from '@testing-library/react';
import { afterEach, beforeAll, describe, expect, it, vi } from 'vitest';
import { setupTestApp } from '../../../test-support';
import { api, house, sentReviews, start } from './studyTestKit';

beforeAll(setupTestApp);

afterEach(() => {
  cleanup();
  vi.unstubAllGlobals();
});

const cloze = {
  ...house,
  id: 'c3',
  item_id: 'i3',
  direction: 'productive',
  state: 'review',
  exercise: 'cloze',
  term: 'house',
  gloss: 'casa',
  cloze_text: 'This is a big ____.',
  example: { text: 'This is a big house.', translation: 'Esta é uma casa grande.' },
  options: undefined,
};
const picture = {
  ...cloze,
  id: 'c4',
  item_id: 'i4',
  exercise: 'image',
  cloze_text: undefined,
  image_url: '/media/openmoji/1F3E0.svg',
};

describe('lacuna', () => {
  it('mostra a frase com a lacuna e a tradução, sem revelar a palavra, e corrige digitando', async () => {
    const calls = api([cloze]);
    await start();
    expect(await screen.findByRole('heading', { name: 'This is a big ____.' })).toBeTruthy();
    expect(screen.getByText('Complete a frase com a palavra que falta')).toBeTruthy();
    expect(screen.getByText('Esta é uma casa grande.')).toBeTruthy();
    expect(screen.queryByText('This is a big house.')).toBeNull();

    fireEvent.change(screen.getByLabelText('Sua resposta em inglês'), { target: { value: 'House' } });
    fireEvent.click(screen.getByRole('button', { name: 'Verificar' }));
    expect(await screen.findByText('Correto!')).toBeTruthy();
    expect(screen.getByText('This is a big house.')).toBeTruthy();
    fireEvent.click(screen.getByRole('button', { name: /Fácil/ }));

    expect(await screen.findByText('Sessão concluída')).toBeTruthy();
    expect(sentReviews(calls)[0]).toMatchObject({
      card_id: 'c3',
      exercise: 'cloze',
      answer: 'House',
      client_correct: true,
    });
  });

  it('errar mostra a palavra certa e pergunta "Como foi?"', async () => {
    api([cloze]);
    await start();
    fireEvent.change(await screen.findByLabelText('Sua resposta em inglês'), { target: { value: 'cat' } });
    fireEvent.click(screen.getByRole('button', { name: 'Verificar' }));
    expect(await screen.findByText('Como foi?')).toBeTruthy();
    expect(screen.getByText('house', { selector: 'strong' })).toBeTruthy();
  });

  it('sem a frase no cartão, cai no digitar a partir da tradução', async () => {
    api([{ ...cloze, cloze_text: undefined }]);
    await start();
    expect(await screen.findByRole('heading', { name: 'casa' })).toBeTruthy();
  });
});

describe('imagem → palavra', () => {
  it('mostra só a imagem (sem a tradução) e a dica revela a 1ª letra', async () => {
    const calls = api([picture]);
    await start();
    const image = await screen.findByRole('img', { name: 'Imagem da palavra que você precisa escrever' });
    expect(image.getAttribute('src')).toBe('/media/openmoji/1F3E0.svg');
    expect(screen.getByText('Que palavra em inglês a imagem mostra?')).toBeTruthy();
    expect(screen.queryByText('casa')).toBeNull();

    fireEvent.click(screen.getByRole('button', { name: 'Dica' }));
    expect(screen.getByText('Começa com “h”')).toBeTruthy();
    fireEvent.change(screen.getByLabelText('Sua resposta em inglês'), { target: { value: 'house' } });
    fireEvent.click(screen.getByRole('button', { name: 'Verificar' }));
    fireEvent.click(await screen.findByRole('button', { name: 'Continuar' }));

    expect(await screen.findByText('Sessão concluída')).toBeTruthy();
    expect(sentReviews(calls)[0]).toMatchObject({ exercise: 'image', hint_used: true, difficulty: 'hard' });
  });

  it('sem imagem no cartão, cai no digitar a partir da tradução', async () => {
    api([{ ...picture, image_url: undefined }]);
    await start();
    expect(await screen.findByRole('heading', { name: 'casa' })).toBeTruthy();
  });
});
