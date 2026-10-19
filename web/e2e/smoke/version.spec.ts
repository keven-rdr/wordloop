import { expect, test } from '@playwright/test';

test('configuracoes mostra a versao do web e da API', async ({ page }) => {
  await page.route('**/api/v1/version', (route) =>
    route.fulfill({ json: { api: '0.1.0', commit: 'abc1234', env: 'tst' } }),
  );
  await page.route('**/api/v1/me', (route) =>
    route.fulfill({ json: { id: 'u1', display_name: 'Maria', role: 'user', two_factor_enabled: false } }),
  );
  await page.goto('/settings');
  await expect(page.getByRole('heading', { name: 'Configurações' })).toBeVisible();
  await expect(page.getByTestId('version-badge')).toContainText('API v0.1.0');
});
