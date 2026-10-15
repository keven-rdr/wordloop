import { expect, type Page, test } from '@playwright/test';

const unitId = '11111111-1111-1111-1111-111111111111';
const unit = { id: unitId, position: 1, theme_key: 'home-and-family', item_count: 2 };
const house = {
  id: 'a1',
  kind: 'word',
  term: 'house',
  gloss: 'casa',
  cefr: 'A1',
  unit_id: unitId,
  status: 'published',
};
const mother = {
  id: 'a2',
  kind: 'word',
  term: 'mother',
  gloss: 'mãe',
  cefr: 'A1',
  unit_id: unitId,
  status: 'published',
};

async function mockApi(page: Page) {
  await page.route('**/api/v1/version', (route) =>
    route.fulfill({ json: { api: '0.1.0', commit: 'abc1234', env: 'tst' } }),
  );
  await page.route('**/api/v1/me', (route) =>
    route.fulfill({
      json: {
        id: '00000000-0000-0000-0000-000000000001',
        display_name: 'Ana',
        role: 'user',
        two_factor_enabled: false,
      },
    }),
  );
  await page.route('**/api/v1/catalog/units', (route) => route.fulfill({ json: [unit] }));
  await page.route('**/api/v1/catalog/items?*', (route) => {
    const q = new URL(route.request().url()).searchParams.get('q');
    const items = q ? [house] : [house, mother];
    return route.fulfill({ json: { items, page: 1, size: 100, total: items.length } });
  });
  await page.route('**/api/v1/catalog/items/a1', (route) =>
    route.fulfill({
      json: {
        ...house,
        examples: [{ text: 'This is a big house.', translation: 'Esta é uma casa grande.' }],
        images: [],
        related: [],
      },
    }),
  );
}

test('da lista de unidades ao baralho: busca, virar o cartão e avançar', async ({ page }) => {
  await mockApi(page);
  await page.goto('/words');
  await expect(page.getByRole('heading', { name: 'Palavras' })).toBeVisible();
  await expect(page.getByRole('link', { name: /Casa e família/ })).toContainText('2 palavras');
  await page.screenshot({ path: 'test-results/words-units.png' });

  await page.getByRole('searchbox').fill('ho');
  await expect(page.getByText('1 resultado')).toBeVisible();
  await page.getByRole('searchbox').fill('');

  await page.getByRole('link', { name: /Casa e família/ }).click();
  await expect(page).toHaveURL(new RegExp(`/words/${unitId}$`));
  await expect(page.getByText('1/2')).toBeVisible();
  await page.screenshot({ path: 'test-results/words-card-front.png' });

  await page.getByRole('button', { name: 'Mostrar a tradução' }).click();
  await expect(page.getByText('Esta é uma casa grande.')).toBeVisible();
  await page.screenshot({ path: 'test-results/words-card-back.png' });

  await page.getByRole('button', { name: 'Próximo cartão' }).click();
  await expect(page.getByText('2/2')).toBeVisible();
  await expect(page.getByText('mother')).toBeVisible();
});
