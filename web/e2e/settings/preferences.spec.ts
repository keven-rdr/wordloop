import { expect, type Page, test } from '@playwright/test';

const defaults = { ui_locale: 'system', theme: 'system', tz: 'America/Sao_Paulo', tz_auto: true };

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
  await page.route('**/api/v1/me/settings', (route) => {
    if (route.request().method() === 'PUT') {
      const body = route.request().postDataJSON() as Record<string, unknown>;
      puts.push(body);
      return route.fulfill({ json: body });
    }
    return route.fulfill({ json: defaults });
  });
  return puts;
}

test('preferências: trocar o idioma e o tema salva e já vale na tela', async ({ page }) => {
  const puts = await mockApi(page);
  await page.goto('/preferences');
  await expect(page.getByRole('heading', { name: 'Preferências' })).toBeVisible();
  await page.screenshot({ path: 'test-results/preferences.png', fullPage: true });

  await page.getByText('English (US)').click();
  await expect(page.getByRole('heading', { name: 'Preferences' })).toBeVisible();
  await expect(page.getByText('Saved')).toBeVisible();

  await page.getByText('Dark', { exact: true }).click();
  await expect.poll(() => puts.length).toBe(2);
  expect(puts[1]).toEqual({ ...defaults, ui_locale: 'en-US', theme: 'dark' });
  expect(await page.evaluate(() => window.localStorage.getItem('wl-theme'))).toBe('dark');
});

test('preferências: escolher um fuso o fixa; os alvos de toque têm 48 px e a tela cabe no aparelho', async ({
  page,
}) => {
  const puts = await mockApi(page);
  await page.goto('/preferences');
  await expect(page.getByRole('heading', { name: 'Preferências' })).toBeVisible();

  for (const name of ['Português (Brasil)', 'Claro', 'Seguir o fuso do aparelho']) {
    const box = await page.locator('label', { hasText: name }).first().boundingBox();
    expect(box?.height ?? 0).toBeGreaterThanOrEqual(48);
  }

  await page.getByText('Seguir o fuso do aparelho').click();
  const select = page.getByLabel('Escolha o fuso');
  await expect(select).toBeVisible();
  expect((await select.boundingBox())?.height ?? 0).toBeGreaterThanOrEqual(48);
  await select.selectOption('Asia/Tokyo');
  await expect.poll(() => puts.length).toBe(2);
  expect(puts[1]).toEqual({ ...defaults, tz: 'Asia/Tokyo', tz_auto: false });

  const overflow = await page.evaluate(
    () => document.documentElement.scrollWidth - document.documentElement.clientWidth,
  );
  expect(overflow).toBeLessThanOrEqual(0);
});
