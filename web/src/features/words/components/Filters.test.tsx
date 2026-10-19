import { cleanup, fireEvent, screen, waitFor } from '@testing-library/react';
import { afterEach, beforeAll, describe, expect, it, vi } from 'vitest';
import { json, renderApp, setupTestApp, stubApi, version } from '../../../test-support';

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

function api() {
  return stubApi({
    'GET /version': () => json(version),
    'GET /me': () => json(me),
    'GET /catalog/units': () => json([unit]),
    'GET /catalog/items': () => json({ items: [house], page: 1, size: 30, total: 1 }),
  });
}

const lastItems = (calls: ReturnType<typeof api>) => calls.filter((c) => c.path === '/catalog/items').at(-1);
const search = async () => {
  const calls = api();
  const { router } = renderApp('/words');
  const units = await screen.findByRole('combobox', { name: 'Filtrar por unidade' });
  await screen.findByRole('option', { name: 'Casa e família' });
  return { calls, router, units };
};

describe('Filtros da busca', () => {
  it('mostra as unidades e os níveis, com tudo selecionado no começo', async () => {
    const { units } = await search();
    expect(units).toHaveValue('');
    expect(screen.getByRole('combobox', { name: 'Filtrar por nível' })).toHaveValue('');
    expect(screen.getByRole('option', { name: 'Todos os níveis' })).toBeInTheDocument();
    expect(screen.getAllByRole('option', { name: /^[ABC][12]$/ })).toHaveLength(6);
  });

  it('escolher uma unidade lista as palavras dela sem digitar nada', async () => {
    const { calls, units } = await search();
    fireEvent.change(units, { target: { value: unit.id } });
    expect(await screen.findByText('1 resultado')).toBeInTheDocument();
    expect(screen.getByText('house')).toBeInTheDocument();
    expect(lastItems(calls)?.method).toBe('GET');
  });
});

describe('Filtros da busca: o que vai para a API', () => {
  it('o nível vai para a API junto com a unidade e o texto', async () => {
    const seen: URL[] = [];
    stubApi({
      'GET /version': () => json(version),
      'GET /me': () => json(me),
      'GET /catalog/units': () => json([unit]),
      'GET /catalog/items': (req) => {
        seen.push(new URL(req.url));
        return json({ items: [house], page: 1, size: 30, total: 1 });
      },
    });
    renderApp('/words');
    const units = await screen.findByRole('combobox', { name: 'Filtrar por unidade' });
    await screen.findByRole('option', { name: 'Casa e família' });
    fireEvent.change(units, { target: { value: unit.id } });
    fireEvent.change(screen.getByRole('combobox', { name: 'Filtrar por nível' }), { target: { value: 'A1' } });
    fireEvent.change(screen.getByRole('searchbox'), { target: { value: 'ho' } });
    await waitFor(() => {
      const last = seen.at(-1);
      expect(last?.searchParams.get('unit')).toBe(unit.id);
      expect(last?.searchParams.get('cefr')).toBe('A1');
      expect(last?.searchParams.get('q')).toBe('ho');
    });
  });

  it('com um filtro ligado, uma letra só não pede mais letras nem manda o texto', async () => {
    const seen: URL[] = [];
    stubApi({
      'GET /version': () => json(version),
      'GET /me': () => json(me),
      'GET /catalog/units': () => json([unit]),
      'GET /catalog/items': (req) => {
        seen.push(new URL(req.url));
        return json({ items: [house], page: 1, size: 30, total: 1 });
      },
    });
    renderApp('/words');
    fireEvent.change(await screen.findByRole('combobox', { name: 'Filtrar por nível' }), { target: { value: 'A1' } });
    fireEvent.change(screen.getByRole('searchbox'), { target: { value: 'h' } });
    expect(await screen.findByText('1 resultado')).toBeInTheDocument();
    expect(screen.queryByText('Digite pelo menos 2 letras para buscar.')).not.toBeInTheDocument();
    expect(seen.at(-1)?.searchParams.has('q')).toBe(false);
  });
});

describe('Filtros da busca: limpar', () => {
  it('voltar a "todas" e apagar o texto volta para a lista de unidades', async () => {
    const { units } = await search();
    fireEvent.change(units, { target: { value: unit.id } });
    await screen.findByText('1 resultado');
    fireEvent.change(units, { target: { value: '' } });
    expect(await screen.findByRole('link', { name: /Casa e família/ })).toBeInTheDocument();
  });

  it('um valor de nível desconhecido é ignorado', async () => {
    await search();
    const level = screen.getByRole('combobox', { name: 'Filtrar por nível' });
    fireEvent.change(level, { target: { value: 'A1' } });
    expect(level).toHaveValue('A1');
    fireEvent.change(level, { target: { value: '' } });
    expect(await screen.findByRole('link', { name: /Casa e família/ })).toBeInTheDocument();
  });
});
