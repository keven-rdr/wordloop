import { expect, type Page, test } from '@playwright/test';

const me = { id: '3f0c4f4e-8d1e-4a3b-9a55-1d2c3b4a5e6f', display_name: 'Maria', role: 'admin' };

type Seen = { login?: unknown; set?: unknown; csrf: (string | undefined)[] };

async function mockApi(page: Page, answer: { login: number; set: number }) {
  const seen: Seen = { csrf: [] };
  let signedIn = false;
  await page.route('**/api/v1/version', (route) =>
    route.fulfill({ json: { api: '0.1.0', commit: 'abc1234', env: 'tst' } }),
  );
  await page.route('**/api/v1/me', (route) =>
    signedIn ? route.fulfill({ json: me }) : route.fulfill({ status: 401 }),
  );
  await page.route('**/api/v1/auth/password/login', (route) => {
    seen.login = route.request().postDataJSON();
    seen.csrf.push(route.request().headers()['x-requested-with']);
    signedIn = answer.login === 204;
    return route.fulfill({ status: answer.login });
  });
  await page.route('**/api/v1/auth/password/set', (route) => {
    seen.set = route.request().postDataJSON();
    seen.csrf.push(route.request().headers()['x-requested-with']);
    signedIn = answer.set === 204;
    return route.fulfill({ status: answer.set });
  });
  return seen;
}

test('login com e-mail e senha: manda ao servidor e mostra o Olá', async ({ page }) => {
  const seen = await mockApi(page, { login: 204, set: 204 });
  await page.goto('/login');
  await page.getByLabel('E-mail').fill('dono@exemplo.com');
  await page.getByLabel('Senha', { exact: true }).fill('uma senha bem comprida');
  await page.getByRole('button', { name: 'Entrar', exact: true }).click();
  await expect(page.getByText('Olá, Maria!')).toBeVisible();
  expect(seen.login).toEqual({ email: 'dono@exemplo.com', password: 'uma senha bem comprida' });
  expect(seen.csrf).toEqual(['XMLHttpRequest']);
  await expect(page).not.toHaveURL(/senha/);
});

test('login com senha errada mostra o aviso e continua na tela', async ({ page }) => {
  await mockApi(page, { login: 401, set: 204 });
  await page.goto('/login');
  await page.getByLabel('E-mail').fill('dono@exemplo.com');
  await page.getByLabel('Senha', { exact: true }).fill('errada errada errada');
  await page.getByRole('button', { name: 'Entrar', exact: true }).click();
  await expect(page.getByRole('alert')).toContainText('E-mail ou senha incorretos');
  await expect(page).toHaveURL(/\/login$/);
});

test('convite: define a senha, esconde o segredo e entra', async ({ page }) => {
  const seen = await mockApi(page, { login: 204, set: 204 });
  await page.goto('/invite#segredo-do-convite');
  await expect(page).toHaveURL(/\/invite$/);
  await page.getByLabel('Nova senha').fill('uma senha bem comprida');
  await page.getByLabel('Repita a senha').fill('uma senha bem comprida');
  await page.getByRole('button', { name: 'Definir senha e entrar' }).click();
  await expect(page.getByText('Olá, Maria!')).toBeVisible();
  expect(seen.set).toEqual({ token: 'segredo-do-convite', password: 'uma senha bem comprida' });
});
