import { expect, test } from '@playwright/test';

const version = { api: '0.1.0', commit: 'abc1234', env: 'tst' };

test.describe('idioma', () => {
  test('brasileiro com en-US na lista de idiomas continua em português', async ({ page }) => {
    await page.addInitScript(() =>
      Object.defineProperty(navigator, 'languages', { get: () => ['pt-BR', 'pt', 'en-US', 'en'] }),
    );
    await page.route('**/api/v1/version', (route) => route.fulfill({ json: version }));
    await page.goto('/');
    await expect(page.locator('main p')).toContainText('vocabulário de inglês');
  });

  test('nunca mostra chaves cruas antes da tradução carregar', async ({ page }) => {
    await page.addInitScript(() => {
      const w = window as unknown as { __raw: string[] };
      w.__raw = [];
      new MutationObserver(() => {
        const text = document.body?.innerText ?? '';
        if (/\b(app|footer)\.[a-zA-Z]+/.test(text)) w.__raw.push(text.slice(0, 80));
      }).observe(document, { childList: true, subtree: true, characterData: true });
    });
    await page.route('**/api/v1/version', (route) => route.fulfill({ json: version }));
    await page.goto('/');
    await expect(page.locator('main p')).toContainText('vocabulário');
    const raw = await page.evaluate(() => (window as unknown as { __raw: string[] }).__raw);
    expect(raw).toEqual([]);
  });
});
