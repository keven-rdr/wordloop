import { expect, type Page, test } from '@playwright/test';

// Montar a frase: a pergunta é a tradução; as peças vêm embaralhadas e sem pontuação do servidor.
const sentence = {
  id: 'c5',
  item_id: 'i5',
  direction: 'productive',
  state: 'review',
  exercise: 'assemble',
  term: 'I like the red house.',
  gloss: 'Eu gosto da casa vermelha.',
  tiles: ['red', 'I', 'house', 'the', 'like'],
  mastery: 0.3,
};

async function mockApi(page: Page) {
  const sent: { reviews: Record<string, unknown>[] }[] = [];
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
  await page.route('**/api/v1/goals/current', (route) =>
    route.fulfill({
      json: {
        goal: {
          id: 'g1',
          period_start: '2026-11-16',
          period_end: '2026-11-22',
          preset: 'steady',
          daily_minutes: 5,
          days_target: 7,
          weekdays: [1, 2, 3, 4, 5, 6, 7],
        },
        ended: false,
      },
    }),
  );
  await page.route('**/api/v1/study/plan', (route) =>
    route.fulfill({
      json: {
        session_id: 'sess-1',
        cards: [sentence],
        new_limit: 8,
        review_limit: 20,
        expires_at: '2999-01-01T00:00:00Z',
      },
    }),
  );
  await page.route('**/api/v1/study/reviews', (route) => {
    const body = route.request().postDataJSON() as { reviews: { id: string }[] };
    sent.push(body as { reviews: Record<string, unknown>[] });
    return route.fulfill({
      json: { results: body.reviews.map((review) => ({ id: review.id, status: 'applied', correct: true })) },
    });
  });
  return sent;
}

const bank = (page: Page) => page.getByRole('list', { name: 'Palavras para montar a frase' });

async function open(page: Page) {
  await page.goto('/study');
  await page.getByRole('button', { name: 'Começar' }).click();
  await expect(page.getByRole('heading', { name: 'Eu gosto da casa vermelha.' })).toBeVisible();
}

test('montar a frase tocando nas peças: alvo de 48 px, sem rolagem horizontal e resposta certa', async ({ page }) => {
  const sent = await mockApi(page);
  await open(page);
  await expect(page.getByText('I like the red house.')).toHaveCount(0); // a resposta não aparece antes
  for (const word of ['red', 'I', 'house', 'the', 'like']) {
    const box = await bank(page).getByRole('button', { name: word, exact: true }).boundingBox();
    expect(box?.height ?? 0, `peça ${word}`).toBeGreaterThanOrEqual(48);
  }
  expect(await page.evaluate(() => document.documentElement.scrollWidth <= window.innerWidth)).toBe(true);
  await expect(page.getByRole('button', { name: 'Verificar' })).toBeDisabled();

  for (const word of ['I', 'like', 'the', 'red', 'house'])
    await bank(page).getByRole('button', { name: word, exact: true }).click();
  await expect(page.getByRole('button', { name: 'Verificar' })).toBeEnabled();
  await page.screenshot({ path: 'test-results/study-assemble.png' });
  await page.getByRole('button', { name: 'Verificar' }).click();
  await expect(page.getByText('Correto!')).toBeVisible();
  await page.getByRole('button', { name: 'Médio' }).click();
  await expect(page.getByText('Sessão concluída')).toBeVisible();
  expect(sent[0]?.reviews[0]).toMatchObject({
    exercise: 'assemble',
    answer: 'I like the red house',
    client_correct: true,
  });
});

test('montar a frase só com o teclado: números, Backspace e Enter', async ({ page }) => {
  const sent = await mockApi(page);
  await open(page);
  await page.keyboard.press('2'); // I
  await page.keyboard.press('Backspace'); // desfaz
  await expect(bank(page).getByRole('button')).toHaveCount(5);
  for (const key of ['2', '4', '3', '1', '1']) await page.keyboard.press(key); // I like the red house
  await page.keyboard.press('Enter');
  await expect(page.getByText('Correto!')).toBeVisible();
  await page.keyboard.press('2');
  await expect(page.getByText('Sessão concluída')).toBeVisible();
  expect(sent[0]?.reviews[0]).toMatchObject({ exercise: 'assemble', answer: 'I like the red house' });
});

test('errar mostra a frase certa e pergunta como foi', async ({ page }) => {
  await mockApi(page);
  await open(page);
  for (const word of ['like', 'I', 'the', 'red', 'house'])
    await bank(page).getByRole('button', { name: word, exact: true }).click();
  await page.getByRole('button', { name: 'Verificar' }).click();
  await expect(page.getByText('Como foi?')).toBeVisible();
  await expect(page.locator('strong', { hasText: 'I like the red house.' })).toBeVisible();
  await page.screenshot({ path: 'test-results/study-assemble-wrong.png' });
});
