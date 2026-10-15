import { expect, type Page, test } from '@playwright/test';

async function mockApi(page: Page, providers: string[]) {
  await page.route('**/api/v1/version', (route) =>
    route.fulfill({ json: { api: '0.1.0', commit: 'abc1234', env: 'tst' } }),
  );
  await page.route('**/api/v1/me', (route) => route.fulfill({ status: 401 }));
  await page.route('**/api/v1/auth/providers', (route) => route.fulfill({ json: { providers } }));
}

test('com Google e GitHub ligados, os botões levam ao servidor e a volta com erro mostra o motivo', async ({
  page,
}) => {
  await mockApi(page, ['google', 'github']);
  await page.route('**/api/v1/auth/login*', (route) => {
    const failures: Record<string, string> = { github: 'email', google: 'failed' };
    const provider = new URL(route.request().url()).searchParams.get('provider') ?? '';
    return route.fulfill({ status: 302, headers: { location: `/login?error=${failures[provider]}` } });
  });
  await page.goto('/login');
  await expect(page.getByRole('link', { name: /Continuar com Google/ })).toHaveAttribute(
    'href',
    '/api/v1/auth/login?provider=google',
  );
  await page.getByRole('link', { name: /Continuar com GitHub/ }).click();
  await expect(page).toHaveURL(/\/login\?error=email$/);
  await expect(page.getByRole('alert')).toContainText('não informou um e-mail confirmado');
});

test('sem provedores ligados os botões ficam desligados com "em breve"', async ({ page }) => {
  await mockApi(page, []);
  await page.goto('/signup');
  await expect(page.getByRole('button', { name: /Continuar com Google/ })).toBeDisabled();
  await expect(page.getByRole('button', { name: /Continuar com GitHub/ })).toBeDisabled();
  await expect(page.getByText('em breve')).toHaveCount(2);
});
