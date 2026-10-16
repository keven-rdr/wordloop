import { cleanup, fireEvent, screen, waitFor } from '@testing-library/react';
import { afterEach, beforeAll, describe, expect, it, vi } from 'vitest';
import { json, renderApp, setupTestApp, status, stubApi, version } from '../../../test-support';

beforeAll(setupTestApp);

afterEach(() => {
  cleanup();
  vi.unstubAllGlobals();
});

const me = { id: '00000000-0000-0000-0000-000000000001', display_name: 'Ana', role: 'user', two_factor_enabled: false };
const unit = { id: '11111111-1111-1111-1111-111111111111', position: 1, theme_key: 'home-and-family', item_count: 1 };
const house = {
  id: 'a1',
  kind: 'word',
  term: 'house',
  gloss: 'casa',
  cefr: 'A1',
  unit_id: unit.id,
  status: 'published',
  image_url: '/media/openmoji/1F3E0.svg',
};
const credit = {
  code: 'openmoji',
  name: 'OpenMoji',
  license: 'CC-BY-SA-4.0',
  notice: 'All emojis designed by OpenMoji',
};

function api(extra: Record<string, () => Response> = {}) {
  return stubApi({
    'GET /version': () => json(version),
    'GET /me': () => json(me),
    'GET /catalog/units': () => json([unit]),
    'GET /catalog/items': () => json({ items: [house], page: 1, size: 100, total: 1 }),
    'GET /catalog/credits': () => json([credit]),
    ...extra,
  });
}

describe('Créditos', () => {
  it('lista as fontes com a licença e o texto de crédito', async () => {
    api();
    renderApp('/words/credits');
    expect(await screen.findByRole('heading', { name: 'Créditos' })).toBeInTheDocument();
    expect(await screen.findByText('OpenMoji')).toBeInTheDocument();
    expect(screen.getByText('CC-BY-SA-4.0')).toBeInTheDocument();
    expect(screen.getByText('All emojis designed by OpenMoji')).toBeInTheDocument();
  });

  it('o link da tela Palavras abre os créditos e o voltar retorna', async () => {
    api();
    const { router } = renderApp('/words');
    fireEvent.click(await screen.findByRole('link', { name: 'Créditos e licenças' }));
    await waitFor(() => expect(router.state.location.pathname).toBe('/words/credits'));
    fireEvent.click(await screen.findByRole('link', { name: 'Voltar às palavras' }));
    await waitFor(() => expect(router.state.location.pathname).toBe('/words'));
  });

  it('avisa quando os créditos não carregam e sem sessão volta para o login', async () => {
    api({ 'GET /catalog/credits': () => status(500) });
    renderApp('/words/credits');
    expect(await screen.findByRole('alert')).toHaveTextContent('Não foi possível carregar agora.');
    cleanup();
    api({ 'GET /me': () => status(401) });
    renderApp('/words/credits');
    expect(await screen.findByRole('heading', { name: 'Entrar' })).toBeInTheDocument();
  });
});

describe('Imagem do cartão', () => {
  it('mostra a imagem na frente do cartão e a esconde ao virar', async () => {
    api({ 'GET /catalog/items/a1': () => json({ ...house, examples: [], images: [], related: [] }) });
    const { container } = renderApp(`/words/${unit.id}`);
    expect(await screen.findByText('house')).toBeInTheDocument();
    expect(container.querySelector('img')).toHaveAttribute('src', '/media/openmoji/1F3E0.svg');
    fireEvent.click(screen.getByRole('button', { name: 'Mostrar a tradução' }));
    expect(await screen.findByText('casa')).toBeInTheDocument();
    expect(container.querySelector('img')).toBeNull();
  });
});
