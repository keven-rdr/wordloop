import { expect, type Page, test } from '@playwright/test';

const suggestions = {
  based_on_history: true,
  suggestions: [
    { preset: 'light', daily_minutes: 5, days_target: 2 },
    { preset: 'steady', daily_minutes: 10, days_target: 3 },
    { preset: 'intense', daily_minutes: 20, days_target: 4 },
  ],
};
const goal = {
  id: 'g1',
  period_start: '2026-10-19',
  period_end: '2026-10-25',
  preset: 'intense',
  daily_minutes: 20,
  days_target: 4,
  weekdays: [1, 2, 3, 4, 5, 6, 7],
};
const started = {
  goal,
  progress: {
    days_studied: 0,
    days_target: 4,
    minutes_studied: 0,
    minutes_goal: 80,
    new_words: 0,
    days_left: 6,
    met: false,
  },
  ended: false,
};

async function mockApi(page: Page, current: unknown) {
  const posts: Record<string, unknown>[] = [];
  let now = current;
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
  await page.route('**/api/v1/progress/dashboard*', (route) =>
    route.fulfill({
      json: {
        today: '2026-10-19',
        streak: 0,
        studied_today: false,
        today_stats: { reviews: 0, correct: 0, new_items: 0, active_seconds: 0 },
        forecast: [],
        known_words: 0,
        most_missed: [],
      },
    }),
  );
  await page.route('**/api/v1/progress/heatmap*', (route) =>
    route.fulfill({ json: { from: '2026-07-28', to: '2026-10-19', days: [] } }),
  );
  await page.route('**/api/v1/goals/suggestions', (route) => route.fulfill({ json: suggestions }));
  await page.route('**/api/v1/goals/current', (route) => route.fulfill({ json: now }));
  await page.route('**/api/v1/goals', (route) => {
    posts.push(route.request().postDataJSON() as Record<string, unknown>);
    now = started;
    return route.fulfill({ status: 201, json: started });
  });
  return posts;
}

test('sem meta: o cartão do Início leva à escolha, e começar volta ao Início com a meta', async ({ page }) => {
  const posts = await mockApi(page, { ended: false });
  await page.goto('/');
  await expect(page.getByText('Defina sua meta da semana')).toBeVisible();
  await page.getByRole('link', { name: 'Definir meta' }).click();

  await expect(page).toHaveURL(/\/goals$/);
  await expect(page.getByText('Sugestões baseadas no seu ritmo das últimas 4 semanas.')).toBeVisible();
  await page.screenshot({ path: 'test-results/goals-choose.png', fullPage: true });
  await page.getByText('Intensa').click();
  await page.getByRole('button', { name: 'Começar' }).click();

  await expect(page).toHaveURL(/\/$/);
  await expect(page.getByText('0 de 4 dias estudados')).toBeVisible();
  await expect(page.getByText('Faltam 6 dias')).toBeVisible();
  await page.screenshot({ path: 'test-results/goals-home.png', fullPage: true });
  expect(posts).toEqual([{ preset: 'intense', daily_minutes: 20, days_target: 4 }]);
});

test('a tela de metas tem alvos de toque de 48 px e cabe na largura do aparelho', async ({ page }) => {
  await mockApi(page, { ended: false });
  await page.goto('/goals');
  await expect(page.getByRole('button', { name: 'Começar' })).toBeVisible();
  for (const name of [/Leve/, /Constante/, /Intensa/, /Personalizar/]) {
    const box = await page.locator('label', { hasText: name }).first().boundingBox();
    expect(box?.height ?? 0).toBeGreaterThanOrEqual(48);
  }
  const start = await page.getByRole('button', { name: 'Começar' }).boundingBox();
  expect(start?.height ?? 0).toBeGreaterThanOrEqual(48);

  await page.getByText('Personalizar').click();
  const field = await page.getByLabel('Minutos por dia').boundingBox();
  expect(field?.height ?? 0).toBeGreaterThanOrEqual(48);

  const overflow = await page.evaluate(
    () => document.documentElement.scrollWidth - document.documentElement.clientWidth,
  );
  expect(overflow).toBeLessThanOrEqual(0);
});
