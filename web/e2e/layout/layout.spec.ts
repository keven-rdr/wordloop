import { expect, type Locator, type Page, test } from '@playwright/test';

// Layout: o mobile (375px) continua com o menu inferior e uma coluna; o desktop (1280px) ganha o menu lateral fixo
// e o Início em grade.
const dashboard = {
  today: '2026-11-23',
  streak: 5,
  studied_today: true,
  accuracy_7d: 0.784,
  today_stats: { reviews: 12, correct: 9, new_items: 3, active_seconds: 140 },
  forecast: [9, 0, 3, 0, 0, 0, 4].map((due, at) => ({ date: `2026-11-${23 + at}`, due })),
  known_words: 42,
  most_missed: [{ item_id: 'i1', term: 'house', gloss: 'casa', misses: 3 }],
};
const heatmap = { from: '2026-09-01', to: '2026-11-23', days: [{ date: '2026-11-23', reviews: 12, correct: 9 }] };
const unit = { id: '11111111-1111-1111-1111-111111111111', position: 1, theme_key: 'home-and-family', item_count: 2 };

async function mockApi(page: Page) {
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
  await page.route('**/api/v1/progress/dashboard*', (route) => route.fulfill({ json: dashboard }));
  await page.route('**/api/v1/progress/heatmap*', (route) => route.fulfill({ json: heatmap }));
  await page.route('**/api/v1/goals/current', (route) => route.fulfill({ json: { ended: false } }));
  await page.route('**/api/v1/catalog/units', (route) => route.fulfill({ json: [unit] }));
}

const box = async (locator: Locator) => {
  const found = await locator.boundingBox();
  if (!found) throw new Error('elemento sem caixa');
  return found;
};

const noHorizontalScroll = async (page: Page) => {
  const overflow = await page.evaluate(
    () => document.documentElement.scrollWidth - document.documentElement.clientWidth,
  );
  expect(overflow).toBeLessThanOrEqual(0);
};

const menu = (page: Page) => page.getByRole('navigation', { name: 'Menu principal' });

test.describe('mobile (375px)', () => {
  test.use({ viewport: { width: 375, height: 812 } });

  test('Início: menu inferior e tudo numa coluna', async ({ page }) => {
    await mockApi(page);
    await page.goto('/');
    await expect(page.getByText('5 dias seguidos')).toBeVisible();
    await expect(menu(page)).toHaveCount(1);
    await expect(menu(page).getByText('wordloop')).toHaveCount(0);
    const nav = await box(menu(page));
    expect(nav.y + nav.height).toBeCloseTo(812, 0);
    expect(nav.width).toBe(375);

    const study = await box(page.getByRole('link', { name: /Estudar agora/ }));
    const goal = await box(page.getByRole('heading', { name: 'Defina sua meta da semana' }));
    const progress = await box(page.getByRole('heading', { name: 'Seu progresso' }));
    expect(goal.y).toBeGreaterThan(study.y);
    expect(progress.y).toBeGreaterThan(goal.y);
    expect(progress.x).toBeLessThan(study.x + study.width / 2);
    await noHorizontalScroll(page);
    await page.screenshot({ path: 'test-results/layout-home-375.png', fullPage: true });
  });

  for (const path of ['/words', '/study']) {
    test(`${path}: menu inferior, sem rolagem horizontal`, async ({ page }) => {
      await mockApi(page);
      await page.goto(path);
      await expect(menu(page)).toBeVisible();
      await expect(menu(page).getByRole('link')).toHaveCount(5); // Perfil, Início, Estudar, Ranking e Palavras
      await expect(menu(page).getByText('wordloop')).toHaveCount(0);
      await noHorizontalScroll(page);
    });
  }
});

test.describe('desktop (1280px)', () => {
  test.use({ viewport: { width: 1280, height: 800 }, isMobile: false, hasTouch: false });

  test('Início: menu lateral fixo e grade (atalhos e meta ao lado do painel)', async ({ page }) => {
    await mockApi(page);
    await page.goto('/');
    await expect(page.getByText('5 dias seguidos')).toBeVisible();
    await expect(menu(page)).toHaveCount(1);
    await expect(menu(page).getByText('wordloop')).toBeVisible();
    const nav = await box(menu(page));
    expect(nav.x).toBe(0);
    expect(nav.height).toBe(800);
    expect(nav.width).toBeLessThan(320);

    const study = await box(page.getByRole('link', { name: /Estudar agora/ }));
    const goal = await box(page.getByRole('heading', { name: 'Defina sua meta da semana' }));
    const progress = await box(page.getByRole('heading', { name: 'Seu progresso' }));
    expect(study.x).toBeGreaterThanOrEqual(nav.width);
    expect(goal.y).toBeGreaterThan(study.y);
    expect(progress.x).toBeGreaterThan(study.x + study.width);
    expect(progress.y).toBeLessThan(goal.y);

    // o painel em duas colunas: sequência e acertos lado a lado
    const streak = await box(page.getByRole('region', { name: /sequência/i }).first());
    const accuracy = await box(page.getByRole('region', { name: /acerto/i }).first());
    expect(accuracy.x).toBeGreaterThan(streak.x + streak.width / 2);
    await noHorizontalScroll(page);
    await page.screenshot({ path: 'test-results/layout-home-1280.png', fullPage: true });
  });

  test('o menu lateral marca a tela atual e navega', async ({ page }) => {
    await mockApi(page);
    await page.goto('/words');
    await expect(menu(page).getByRole('link', { name: 'Palavras' })).toHaveAttribute('aria-current', 'page');
    await menu(page).getByRole('link', { name: 'Estudar' }).click();
    await expect(page).toHaveURL(/\/study$/);
    await expect(menu(page).getByRole('link', { name: 'Estudar' })).toHaveAttribute('aria-current', 'page');
    await page.screenshot({ path: 'test-results/layout-study-1280.png', fullPage: true });
  });

  for (const path of ['/words', '/study']) {
    test(`${path}: o conteúdo começa depois do menu lateral`, async ({ page }) => {
      await mockApi(page);
      await page.goto(path);
      const nav = await box(menu(page));
      const heading = await box(page.getByRole('heading', { level: 1 }).first());
      expect(heading.x).toBeGreaterThanOrEqual(nav.width);
      await noHorizontalScroll(page);
    });
  }
});
