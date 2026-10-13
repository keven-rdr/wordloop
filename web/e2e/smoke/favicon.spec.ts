import { expect, test } from '@playwright/test';

test('o favicon da marca e declarado e servido', async ({ page, request }) => {
  await page.goto('/');
  const hrefs = await page
    .locator('link[rel~="icon"], link[rel="apple-touch-icon"]')
    .evaluateAll((els) => els.map((el) => el.getAttribute('href')));
  expect(hrefs).toEqual(['/favicon.svg', '/favicon-32.png', '/favicon-16.png', '/apple-touch-icon.png']);
  for (const href of hrefs) {
    const res = await request.get(href ?? '');
    expect(res.ok(), href ?? '').toBe(true);
    expect(res.headers()['content-type']).toMatch(/^image\//);
  }
});
