import { expect, type Page, test } from '@playwright/test';

const week = {
  period: 'week',
  from: '2026-11-23',
  to: '2026-11-29',
  entries: [
    { position: 1, nickname: 'bia', points: 420, me: false },
    { position: 2, nickname: 'caio_com_apelido_long', points: 300, me: false },
    { position: 3, nickname: 'duda', points: 150, me: false },
    { position: 4, nickname: 'eva', points: 90, me: false },
  ],
  me: { position: 6, points: 40, me: true },
};

async function mockApi(page: Page) {
  const puts: Record<string, unknown>[] = [];
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
  await page.route('**/api/v1/leaderboard?*', (route) => route.fulfill({ json: week }));
  await page.route('**/api/v1/leaderboard/profile', (route) => {
    if (route.request().method() === 'PUT') {
      puts.push(route.request().postDataJSON() as Record<string, unknown>);
      return route.fulfill({ json: { nickname: 'ana_1', visible: true } });
    }
    return route.fulfill({ json: { nickname: '', visible: false } });
  });
  return puts;
}

test('ranking no celular: pódio, minha linha, alvos de 48 px e sem rolagem horizontal', async ({ page }) => {
  const puts = await mockApi(page);
  await page.goto('/ranking');
  await expect(page.getByRole('list', { name: 'Pódio' })).toBeVisible();
  await expect(page.getByRole('region', { name: 'Sua posição' })).toContainText('6º');
  for (const name of ['Semana', 'Geral', 'Salvar']) {
    const box = await page.getByRole('button', { name }).boundingBox();
    expect(box?.height ?? 0).toBeGreaterThanOrEqual(48);
  }
  const overflow = await page.evaluate(
    () => document.documentElement.scrollWidth - document.documentElement.clientWidth,
  );
  expect(overflow).toBeLessThanOrEqual(0);
  await page.screenshot({ path: 'test-results/ranking.png', fullPage: true });

  await page.getByLabel('Apelido').fill('ana_1');
  await page.getByRole('button', { name: 'Salvar' }).click();
  await expect(page.getByText('Salvo!')).toBeVisible();
  expect(puts).toEqual([{ nickname: 'ana_1', visible: true }]);
});
