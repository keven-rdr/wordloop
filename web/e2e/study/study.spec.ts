import { expect, type Page, test } from '@playwright/test';

const choose = {
  id: 'c1',
  item_id: 'i1',
  direction: 'receptive',
  state: 'new',
  exercise: 'choose',
  term: 'house',
  gloss: 'casa',
  mastery: 0,
  options: [
    { item_id: 'i2', gloss: 'mãe' },
    { item_id: 'i1', gloss: 'casa' },
    { item_id: 'i3', gloss: 'gato' },
  ],
};
const typed = { ...choose, id: 'c2', item_id: 'i2', exercise: 'type', term: 'mother', gloss: 'mãe', options: [] };

async function mockApi(page: Page, cards: unknown[]) {
  const sent: { reviews: Record<string, unknown>[] }[] = [];
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
  await page.route('**/api/v1/study/plan', (route) =>
    route.fulfill({
      json: { session_id: 'sess-1', cards, new_limit: 8, review_limit: 20, expires_at: '2026-11-19T04:00:00Z' },
    }),
  );
  await page.route('**/api/v1/study/reviews', (route) => {
    const body = route.request().postDataJSON() as { reviews: { id: string }[] };
    sent.push(body as { reviews: Record<string, unknown>[] });
    return route.fulfill({
      json: { results: body.reviews.map((review) => ({ id: review.id, status: 'applied', correct: true })) },
    });
  });
  return sent;
}

async function expectTouchTarget(page: Page, name: string | RegExp) {
  const box = await page.getByRole('button', { name }).first().boundingBox();
  expect(box?.height ?? 0).toBeGreaterThanOrEqual(48);
}

test('exercício de escolher: verificar, dificuldade em um toque e resumo', async ({ page }) => {
  const sent = await mockApi(page, [choose]);
  await page.goto('/study');
  await expect(page.getByRole('heading', { name: 'Estudar' })).toBeVisible();
  await expectTouchTarget(page, '5 minutos');
  await page.getByRole('button', { name: 'Começar' }).click();

  await expect(page.getByRole('heading', { name: 'house' })).toBeVisible();
  await expectTouchTarget(page, /casa/);
  await page.getByRole('button', { name: /casa/ }).click();
  await page.getByRole('button', { name: 'Verificar' }).click();
  await expect(page.getByText('Correto!')).toBeVisible();
  await expectTouchTarget(page, 'Médio');
  await page.screenshot({ path: 'test-results/study-choose.png' });

  await page.getByRole('button', { name: 'Médio' }).click();
  await expect(page.getByText('Sessão concluída')).toBeVisible();
  await expect(page.getByText('+1 palavra conhecida')).toBeVisible();
  expect(sent).toHaveLength(1);
  expect(sent[0]?.reviews[0]).toMatchObject({ card_id: 'c1', chosen_item_id: 'i1', difficulty: 'medium' });
});

test('exercício de digitar: errar mostra a resposta, pergunta "Como foi?" e o cartão volta', async ({ page }) => {
  await mockApi(page, [typed]);
  await page.goto('/study');
  await page.getByRole('button', { name: 'Começar' }).click();

  const input = page.getByLabel('Sua resposta em inglês');
  await expect(input).toBeFocused();
  await expect(input).toHaveAttribute('autocapitalize', 'none');
  await expect(input).toHaveAttribute('spellcheck', 'false');
  await input.fill('xyz');
  await expectTouchTarget(page, 'Verificar');
  await page.getByRole('button', { name: 'Verificar' }).click();

  await expect(page.getByText('Quase!')).toBeVisible();
  await expect(page.locator('strong', { hasText: 'mother' })).toBeVisible();
  await page.screenshot({ path: 'test-results/study-type-wrong.png' });
  await page.getByRole('button', { name: 'Foi distração' }).click();

  await expect(page.getByLabel('Sua resposta em inglês')).toBeVisible();
  await expect(page.getByLabel('Sua resposta em inglês')).toHaveValue('');
});

test('sair pela ✕ pede confirmação e volta ao plano', async ({ page }) => {
  await mockApi(page, [choose, typed]);
  await page.goto('/study');
  await page.getByRole('button', { name: 'Começar' }).click();
  await page.getByRole('button', { name: 'Sair da sessão' }).click();
  await expect(page.getByRole('alertdialog')).toBeVisible();
  await page.getByRole('button', { name: 'Sair', exact: true }).click();
  await expect(page.getByRole('button', { name: 'Começar' })).toBeVisible();
});
