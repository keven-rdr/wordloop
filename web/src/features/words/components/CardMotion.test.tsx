import { act, cleanup, fireEvent, screen } from '@testing-library/react';
import { afterEach, beforeAll, describe, expect, it, vi } from 'vitest';
import { json, renderApp, setupTestApp, stubApi, version } from '../../../test-support';
import { CARD_MS } from './motion';

beforeAll(setupTestApp);

afterEach(() => {
  cleanup();
  vi.useRealTimers();
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
const mother = {
  id: 'a2',
  kind: 'word',
  term: 'mother',
  gloss: 'mãe',
  cefr: 'A1',
  unit_id: unit.id,
  status: 'published',
};

function api() {
  stubApi({
    'GET /version': () => json(version),
    'GET /me': () => json(me),
    'GET /catalog/units': () => json([unit]),
    'GET /catalog/items': () => json({ items: [house, mother], page: 1, size: 100, total: 2 }),
  });
}

const leaving = () => document.querySelector('[inert]');

describe('Troca de cartão', () => {
  it('o cartão anterior sai por cima do novo e some quando a animação termina', async () => {
    api();
    renderApp(`/words/${unit.id}`);
    expect(await screen.findByText('house')).toBeInTheDocument();
    expect(leaving()).toBeNull();

    vi.useFakeTimers();
    fireEvent.click(screen.getByRole('button', { name: 'Próximo cartão' }));
    expect(screen.getByText('mother')).toBeInTheDocument();
    expect(leaving()).toHaveTextContent('house');

    act(() => {
      vi.advanceTimersByTime(CARD_MS + 10);
    });
    expect(leaving()).toBeNull();
    expect(screen.queryByText('house')).not.toBeInTheDocument();
  });

  it('com movimento reduzido troca direto, sem cartão saindo', async () => {
    api();
    vi.stubGlobal('matchMedia', () => ({
      matches: true,
      addEventListener: () => undefined,
      removeEventListener: () => undefined,
    }));
    renderApp(`/words/${unit.id}`);
    expect(await screen.findByText('house')).toBeInTheDocument();
    fireEvent.click(screen.getByRole('button', { name: 'Próximo cartão' }));
    expect(await screen.findByText('mother')).toBeInTheDocument();
    expect(leaving()).toBeNull();
  });

  it('ignora outro clique enquanto o cartão ainda está saindo', async () => {
    api();
    renderApp(`/words/${unit.id}`);
    expect(await screen.findByText('house')).toBeInTheDocument();
    vi.useFakeTimers();
    fireEvent.click(screen.getByRole('button', { name: 'Próximo cartão' }));
    fireEvent.click(screen.getByRole('button', { name: 'Cartão anterior' }));
    expect(screen.getByText('2/2')).toBeInTheDocument();
  });
});
