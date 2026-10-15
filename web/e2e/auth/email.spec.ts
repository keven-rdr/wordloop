import { expect, type Page, test } from '@playwright/test';

const me = { id: '3f0c4f4e-8d1e-4a3b-9a55-1d2c3b4a5e6f', display_name: 'ana', role: 'user' };

type Seen = { register?: unknown; forgot?: unknown; set?: unknown };

async function mockApi(page: Page) {
  const seen: Seen = {};
  let signedIn = false;
  await page.route('**/api/v1/version', (route) =>
    route.fulfill({ json: { api: '0.1.0', commit: 'abc1234', env: 'tst' } }),
  );
  await page.route('**/api/v1/me', (route) =>
    signedIn ? route.fulfill({ json: me }) : route.fulfill({ status: 401 }),
  );
  await page.route('**/api/v1/auth/register', (route) => {
    seen.register = route.request().postDataJSON();
    return route.fulfill({ status: 202 });
  });
  await page.route('**/api/v1/auth/password/forgot', (route) => {
    seen.forgot = route.request().postDataJSON();
    return route.fulfill({ status: 202 });
  });
  await page.route('**/api/v1/auth/password/set', (route) => {
    seen.set = route.request().postDataJSON();
    signedIn = true;
    return route.fulfill({ status: 204 });
  });
  return seen;
}

test('cadastro: pede o link por e-mail e mostra para onde ele foi', async ({ page }) => {
  const seen = await mockApi(page);
  await page.goto('/signup');
  await page.getByLabel('E-mail').fill('ana@exemplo.com');
  await page.getByRole('button', { name: 'Continuar', exact: true }).click();
  await expect(page.getByRole('status')).toContainText('Enviamos um link para ana@exemplo.com');
  expect(seen.register).toEqual({ email: 'ana@exemplo.com', locale: 'pt-BR' });
});

test('esqueci a senha: o login leva à tela e o pedido não revela se o e-mail existe', async ({ page }) => {
  const seen = await mockApi(page);
  await page.goto('/login');
  await page.getByRole('link', { name: 'Esqueceu a senha?' }).click();
  await expect(page).toHaveURL(/\/forgot$/);
  await page.getByLabel('E-mail').fill('qualquer@exemplo.com');
  await page.getByRole('button', { name: 'Enviar o link' }).click();
  await expect(page.getByRole('status')).toContainText('Se existir uma conta');
  expect(seen.forgot).toEqual({ email: 'qualquer@exemplo.com', locale: 'pt-BR' });
});

test('link do e-mail: esconde o segredo, define a senha e entra', async ({ page }) => {
  const seen = await mockApi(page);
  await page.goto('/verify#segredo-do-e-mail');
  await expect(page).toHaveURL(/\/verify$/);
  await page.getByLabel('Nova senha').fill('uma senha bem comprida');
  await page.getByLabel('Repita a senha').fill('uma senha bem comprida');
  await page.getByRole('button', { name: 'Definir senha e entrar' }).click();
  await expect(page.getByText('Olá, ana!')).toBeVisible();
  expect(seen.set).toEqual({ token: 'segredo-do-e-mail', password: 'uma senha bem comprida' });
});
