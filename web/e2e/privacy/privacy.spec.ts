import { expect, type Page, test } from '@playwright/test';

const exported = { schema_version: 1, profile: { display_name: 'Ana' }, reviews: [] };

async function mockApi(page: Page, eraseStatus = 204) {
  const erased: Record<string, unknown>[] = [];
  let signedIn = true;
  await page.route('**/api/v1/version', (route) =>
    route.fulfill({ json: { api: '0.1.0', commit: 'abc1234', env: 'tst' } }),
  );
  await page.route('**/api/v1/me', (route) =>
    signedIn
      ? route.fulfill({
          json: {
            id: '00000000-0000-0000-0000-000000000001',
            display_name: 'Ana',
            role: 'user',
            two_factor_enabled: false,
          },
        })
      : route.fulfill({ status: 401 }),
  );
  await page.route('**/api/v1/me/settings', (route) => route.fulfill({ status: 401 }));
  await page.route('**/api/v1/auth/providers', (route) => route.fulfill({ json: { providers: [] } }));
  await page.route('**/api/v1/auth/logout', (route) => {
    signedIn = false;
    return route.fulfill({ status: 204 });
  });
  await page.route('**/api/v1/me/export', (route) =>
    route.fulfill({
      json: exported,
      headers: { 'Content-Disposition': 'attachment; filename="wordloop-dados-2026-11-18.json"' },
    }),
  );
  await page.route('**/api/v1/me/erase', (route) => {
    erased.push(route.request().postDataJSON() as Record<string, unknown>);
    return route.fulfill({ status: eraseStatus });
  });
  return erased;
}

test('privacidade e termos abrem sem entrar, cabem na tela e levam um ao outro', async ({ page }) => {
  await mockApi(page);
  await page.route('**/api/v1/me', (route) => route.fulfill({ status: 401 })); // sem sessão
  await page.goto('/privacy');
  await expect(page.getByRole('heading', { name: 'Política de privacidade', level: 1 })).toBeVisible();
  await expect(page.getByRole('note')).toContainText('Rascunho em revisão');
  await expect(page.getByRole('heading', { level: 2 })).toHaveCount(9);
  const overflow = await page.evaluate(
    () => document.documentElement.scrollWidth - document.documentElement.clientWidth,
  );
  expect(overflow).toBeLessThanOrEqual(0);
  await page.screenshot({ path: 'test-results/privacy-policy.png', fullPage: true });

  await page.getByRole('link', { name: 'Termos de uso' }).click();
  await expect(page).toHaveURL(/\/terms$/);
  await expect(page.getByRole('heading', { name: 'Termos de uso', level: 1 })).toBeVisible();
});

test('login e cadastro mostram os links dos termos e da privacidade', async ({ page }) => {
  await mockApi(page);
  await page.route('**/api/v1/me', (route) => route.fulfill({ status: 401 }));
  await page.goto('/login');
  await expect(page.getByText('Ao continuar, você aceita')).toBeVisible();
  await page.getByRole('link', { name: 'Política de privacidade' }).click();
  await expect(page).toHaveURL(/\/privacy$/);
});

test('Seus dados: baixa o arquivo JSON de verdade', async ({ page }) => {
  await mockApi(page);
  await page.goto('/account/data');
  await expect(page.getByRole('heading', { name: 'Seus dados' })).toBeVisible();
  const box = await page.getByRole('button', { name: 'Baixar meus dados' }).boundingBox();
  expect(box?.height ?? 0).toBeGreaterThanOrEqual(48);

  const downloading = page.waitForEvent('download');
  await page.getByRole('button', { name: 'Baixar meus dados' }).click();
  const file = await downloading;
  expect(file.suggestedFilename()).toMatch(/^wordloop-dados-\d{4}-\d{2}-\d{2}\.json$/);
  await expect(page.getByText('Pronto! O arquivo foi baixado.')).toBeVisible();
});

test('Seus dados: excluir a conta pede o e-mail, apaga e volta ao login', async ({ page }) => {
  const erased = await mockApi(page);
  await page.goto('/account/data');
  await page.getByRole('button', { name: 'Excluir minha conta…' }).click();
  const dialog = page.getByRole('alertdialog');
  await expect(dialog).toBeVisible();
  await expect(page.getByRole('button', { name: 'Cancelar' })).toBeFocused();
  await page.screenshot({ path: 'test-results/privacy-erase.png' });

  const confirm = page.getByRole('button', { name: 'Excluir para sempre' });
  await expect(confirm).toBeDisabled();
  expect((await confirm.boundingBox())?.height ?? 0).toBeGreaterThanOrEqual(48);
  await page.getByLabel('E-mail da conta').fill('ana@example.com');
  await confirm.click();

  await expect(page).toHaveURL(/\/login$/);
  expect(erased).toEqual([{ confirmation: 'ana@example.com' }]);
});

test('Seus dados: login antigo mostra o aviso e deixa sair para entrar de novo', async ({ page }) => {
  await mockApi(page, 428);
  await page.goto('/account/data');
  await page.getByRole('button', { name: 'Excluir minha conta…' }).click();
  await page.getByLabel('E-mail da conta').fill('ana@example.com');
  await page.getByRole('button', { name: 'Excluir para sempre' }).click();
  await expect(page.getByText('Por segurança, entre de novo e repita o pedido.')).toBeVisible();
  await page.getByRole('button', { name: 'Sair e entrar de novo' }).click();
  await expect(page).toHaveURL(/\/login$/);
});
