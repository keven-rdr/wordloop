import { expect, test } from '@playwright/test';

const me = { id: '3f0c4f4e-8d1e-4a3b-9a55-1d2c3b4a5e6f', display_name: 'Maria', role: 'user' };

test.beforeEach(async ({ page }) => {
  await page.route('**/api/v1/version', (route) =>
    route.fulfill({ json: { api: '0.1.0', commit: 'abc1234', env: 'tst' } }),
  );
});

test('sem sessão mostra Entrar e leva ao login da API', async ({ page }) => {
  await page.route('**/api/v1/me', (route) => route.fulfill({ status: 401 }));
  await page.route('**/api/v1/auth/login*', (route) =>
    route.fulfill({ status: 200, contentType: 'text/html', body: 'provedor' }),
  );
  await page.goto('/');
  await page.getByRole('link', { name: 'Entrar' }).click();
  await expect(page).toHaveURL(/\/api\/v1\/auth\/login\?return_to=\//);
});

test('com sessão mostra Olá e Sair volta para Entrar', async ({ page }) => {
  let loggedIn = true;
  await page.route('**/api/v1/me', (route) =>
    loggedIn ? route.fulfill({ json: me }) : route.fulfill({ status: 401 }),
  );
  await page.route('**/api/v1/auth/logout', (route) => {
    expect(route.request().method()).toBe('POST');
    expect(route.request().headers()['x-requested-with']).toBe('XMLHttpRequest');
    loggedIn = false;
    return route.fulfill({ status: 204 });
  });
  await page.goto('/');
  await expect(page.getByText('Olá, Maria!')).toBeVisible();
  await expect(page.getByTestId('version-footer')).toContainText('API v0.1.0 · tst');
  await page.getByRole('button', { name: 'Sair' }).click();
  await expect(page.getByRole('link', { name: 'Entrar' })).toBeVisible();
});
