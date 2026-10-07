import { expect, test } from '@playwright/test';

test('rodape mostra a versao do web e da API', async ({ page }) => {
  await page.route('**/api/v1/version', (route) =>
    route.fulfill({ json: { api: '0.1.0', commit: 'abc1234', env: 'tst' } }),
  );
  await page.goto('/');
  await expect(page.getByRole('heading', { name: 'wordloop' })).toBeVisible();
  await expect(page.getByTestId('version-footer')).toContainText('API v0.1.0 · tst');
});
