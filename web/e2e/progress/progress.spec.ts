import { expect, type Page, test } from '@playwright/test';

const dashboard = {
  today: '2026-10-19',
  streak: 5,
  studied_today: true,
  accuracy_7d: 0.784,
  today_stats: { reviews: 12, correct: 9, new_items: 3, active_seconds: 140 },
  forecast: [9, 0, 3, 0, 0, 0, 4].map((due, at) => ({ date: `2026-10-${19 + at}`, due })),
  known_words: 42,
  most_missed: [{ item_id: 'i1', term: 'house', gloss: 'casa', misses: 3 }],
};
const heatmap = {
  from: '2026-07-28',
  to: '2026-10-19',
  days: [
    { date: '2026-10-17', reviews: 31, correct: 25 },
    { date: '2026-10-18', reviews: 10, correct: 8 },
    { date: '2026-10-19', reviews: 12, correct: 9 },
  ],
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
  await page.route('**/api/v1/progress/dashboard*', (route) => route.fulfill({ json: dashboard }));
  await page.route('**/api/v1/progress/heatmap*', (route) => route.fulfill({ json: heatmap }));
}

test('Início: o painel mostra sequência, acertos, heatmap, previsão e palavras mais erradas', async ({ page }) => {
  await mockApi(page);
  await page.goto('/');

  await expect(page.getByRole('heading', { name: 'Seu progresso' })).toBeVisible();
  await expect(page.getByText('5 dias seguidos')).toBeVisible();
  await expect(page.getByText('78%')).toBeVisible();
  await expect(page.getByText('42', { exact: true })).toBeVisible();
  await expect(page.getByRole('img', { name: /Você estudou em 3 dias/ })).toBeVisible();
  await expect(page.getByText('16 revisões nos próximos 7 dias')).toBeVisible();
  await expect(page.getByText('house')).toBeVisible();
  await page.screenshot({ path: 'test-results/progress-home.png', fullPage: true });

  // o atalho principal continua à vista e leva ao estudo
  await page.getByRole('link', { name: /Estudar agora/ }).click();
  await expect(page).toHaveURL(/\/study$/);
});

test('Início: o painel cabe na largura do aparelho, sem rolagem horizontal', async ({ page }) => {
  await mockApi(page);
  await page.goto('/');
  await expect(page.getByText('5 dias seguidos')).toBeVisible();
  const overflow = await page.evaluate(
    () => document.documentElement.scrollWidth - document.documentElement.clientWidth,
  );
  expect(overflow).toBeLessThanOrEqual(0);
  const grid = page.getByRole('img', { name: /Você estudou em/ });
  const box = await grid.boundingBox();
  expect(box?.width ?? 0).toBeGreaterThan(200);
  expect((box?.width ?? 0) <= (page.viewportSize()?.width ?? 0)).toBe(true);
});
