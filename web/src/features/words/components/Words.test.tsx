import { cleanup, fireEvent, screen, waitFor } from '@testing-library/react';
import { afterEach, beforeAll, describe, expect, it, vi } from 'vitest';
import { json, renderApp, setupTestApp, status, stubApi, version } from '../../../test-support';

beforeAll(setupTestApp);

afterEach(() => {
  cleanup();
  vi.unstubAllGlobals();
});

const me = { id: '00000000-0000-0000-0000-000000000001', display_name: 'Ana', role: 'user', two_factor_enabled: false };
const unit = { id: '11111111-1111-1111-1111-111111111111', position: 1, theme_key: 'home-and-family', item_count: 2 };
const house = {
  id: 'a1',
  kind: 'word',
  term: 'house',
  gloss: 'casa',
  cefr: 'A1',
  unit_id: unit.id,
  status: 'published',
};
const mother = { id: 'a2', kind: 'word', term: 'mother', gloss: 'mãe', cefr: 'A1', unit_id: unit.id, status: 'draft' };
const detail = (id: string) => ({
  ...(id === 'a1' ? house : mother),
  examples: [{ text: 'This is a big house.', translation: 'Esta é uma casa grande.' }],
  images: [],
  related: [],
});

function api(extra: Record<string, () => Response> = {}) {
  return stubApi({
    'GET /version': () => json(version),
    'GET /me': () => json(me),
    'GET /catalog/units': () => json([unit]),
    'GET /catalog/items': (req) => {
      const url = new URL(req.url);
      const q = url.searchParams.get('q');
      if (q === 'zzz') return json({ items: [], page: 1, size: 30, total: 0 });
      const items = q ? [house] : [house, mother];
      return json({ items, page: 1, size: 100, total: items.length });
    },
    'GET /catalog/items/a1': () => json(detail('a1')),
    'GET /catalog/items/a2': () => json(detail('a2')),
    ...extra,
  });
}

describe('Tela Palavras', () => {
  it('sem sessão volta para o login', async () => {
    api({ 'GET /me': () => status(401) });
    renderApp('/words');
    expect(await screen.findByRole('heading', { name: 'Entrar' })).toBeInTheDocument();
  });

  it('lista as unidades com o tema traduzido e a contagem', async () => {
    api();
    renderApp('/words');
    const link = await screen.findByRole('link', { name: /Casa e família/ });
    expect(link).toHaveTextContent('2 palavras');
    expect(link).toHaveAttribute('href', `/words/${unit.id}`);
  });

  it('mostra aviso e tenta de novo quando as unidades não carregam', async () => {
    let fail = true;
    api({ 'GET /catalog/units': () => (fail ? status(500) : json([unit])) });
    renderApp('/words');
    expect(await screen.findByRole('alert')).toHaveTextContent('Não foi possível carregar');
    fail = false;
    fireEvent.click(screen.getByRole('button', { name: 'Tentar de novo' }));
    expect(await screen.findByRole('link', { name: /Casa e família/ })).toBeInTheDocument();
  });

  it('sem unidades diz que ainda não há nada', async () => {
    api({ 'GET /catalog/units': () => json([]) });
    renderApp('/words');
    expect(await screen.findByText('Nenhuma unidade disponível para você ainda.')).toBeInTheDocument();
  });
});

describe('Busca', () => {
  it('pede pelo menos 2 letras, mostra resultados e o selo de rascunho', async () => {
    const calls = api({ 'GET /catalog/items': () => json({ items: [house, mother], page: 1, size: 30, total: 2 }) });
    renderApp('/words');
    const box = await screen.findByRole('searchbox', { name: 'Buscar uma palavra ou tradução' });
    fireEvent.change(box, { target: { value: 'h' } });
    expect(await screen.findByText('Digite pelo menos 2 letras para buscar.')).toBeInTheDocument();
    fireEvent.change(box, { target: { value: 'ho' } });
    expect(await screen.findByText('2 resultados')).toBeInTheDocument();
    expect(screen.getByText('house')).toBeInTheDocument();
    expect(screen.getByText('rascunho')).toBeInTheDocument();
    expect(calls.some((c) => c.path === '/catalog/items')).toBe(true);
  });

  it('avisa quando nada é encontrado e quando a busca falha', async () => {
    api();
    renderApp('/words');
    const box = await screen.findByRole('searchbox');
    fireEvent.change(box, { target: { value: 'zzz' } });
    expect(await screen.findByText('Nada encontrado para essa busca.')).toBeInTheDocument();
  });

  it('mostra erro quando a busca falha', async () => {
    api({ 'GET /catalog/items': () => status(500) });
    renderApp('/words');
    fireEvent.change(await screen.findByRole('searchbox'), { target: { value: 'ho' } });
    expect(await screen.findByRole('alert')).toHaveTextContent('Não foi possível carregar');
  });

  it('apagar o texto volta para as unidades', async () => {
    api();
    renderApp('/words');
    const box = await screen.findByRole('searchbox');
    fireEvent.change(box, { target: { value: 'ho' } });
    await screen.findByText('1 resultado');
    fireEvent.change(box, { target: { value: '' } });
    expect(await screen.findByRole('link', { name: /Casa e família/ })).toBeInTheDocument();
  });
});

describe('Baralho de uma unidade', () => {
  it('vira o cartão, mostra exemplo e navega entre as palavras', async () => {
    api();
    renderApp(`/words/${unit.id}`);
    expect(await screen.findByRole('heading', { name: 'Casa e família' })).toBeInTheDocument();
    expect(await screen.findByText('house')).toBeInTheDocument();
    expect(screen.getByText('1/2')).toBeInTheDocument();
    expect(screen.getByRole('button', { name: 'Cartão anterior' })).toBeDisabled();

    fireEvent.click(screen.getByRole('button', { name: 'Mostrar a tradução' }));
    expect(await screen.findByText('casa')).toBeInTheDocument();
    expect(await screen.findByText('Esta é uma casa grande.')).toBeInTheDocument();

    fireEvent.click(screen.getByRole('button', { name: 'Próximo cartão' }));
    expect(await screen.findByText('mother')).toBeInTheDocument();
    expect(screen.getByText('2/2')).toBeInTheDocument();
    expect(screen.getByText('rascunho')).toBeInTheDocument();
    expect(screen.getByRole('button', { name: 'Próximo cartão' })).toBeDisabled();
    expect(screen.getByRole('progressbar')).toHaveAttribute('aria-valuenow', '2');

    fireEvent.click(screen.getByRole('button', { name: 'Cartão anterior' }));
    expect(await screen.findByText('house')).toBeInTheDocument();
    expect(screen.queryByText('casa')).not.toBeInTheDocument();
  });

  it('tocar no cartão também vira e desvira', async () => {
    api();
    renderApp(`/words/${unit.id}`);
    const card = await screen.findByRole('button', { name: /house/ });
    fireEvent.click(card);
    await waitFor(() => expect(screen.getByRole('button', { name: /casa/ })).toHaveAttribute('aria-pressed', 'true'));
    fireEvent.click(screen.getByRole('button', { name: 'Esconder a tradução' }));
    expect(await screen.findByText('Toque no cartão para ver a tradução')).toBeInTheDocument();
  });

  it('unidade vazia, erro e sem sessão', async () => {
    api({ 'GET /catalog/items': () => json({ items: [], page: 1, size: 100, total: 0 }) });
    renderApp(`/words/${unit.id}`);
    expect(await screen.findByText('Esta unidade ainda não tem palavras para você.')).toBeInTheDocument();
    cleanup();

    api({ 'GET /catalog/items': () => status(500) });
    renderApp(`/words/${unit.id}`);
    expect(await screen.findByRole('alert')).toHaveTextContent('Não foi possível carregar');
    cleanup();

    api({ 'GET /me': () => status(401) });
    renderApp(`/words/${unit.id}`);
    expect(await screen.findByRole('heading', { name: 'Entrar' })).toBeInTheDocument();
  });

  it('o botão voltar leva à lista de unidades', async () => {
    api();
    const { router } = renderApp(`/words/${unit.id}`);
    fireEvent.click(await screen.findByRole('link', { name: 'Voltar às unidades' }));
    await waitFor(() => expect(router.state.location.pathname).toBe('/words'));
  });
});

describe('Início', () => {
  it('tem o atalho para Palavras', async () => {
    api();
    renderApp('/');
    expect(await screen.findByRole('link', { name: 'Palavras' })).toHaveAttribute('href', '/words');
  });
});
