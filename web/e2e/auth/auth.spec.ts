import { randomBytes } from 'node:crypto';
import { expect, type Page, test } from '@playwright/test';

const me = { id: '3f0c4f4e-8d1e-4a3b-9a55-1d2c3b4a5e6f', display_name: 'Maria', role: 'user' };
const b64 = (n: number) => randomBytes(n).toString('base64url');

async function addVirtualAuthenticator(page: Page) {
  const cdp = await page.context().newCDPSession(page);
  await cdp.send('WebAuthn.enable');
  await cdp.send('WebAuthn.addVirtualAuthenticator', {
    options: {
      protocol: 'ctap2',
      transport: 'internal',
      hasResidentKey: true,
      hasUserVerification: true,
      isUserVerified: true,
      automaticPresenceSimulation: true,
    },
  });
}

const creationOptions = () => ({
  options: {
    rp: { id: 'localhost', name: 'wordloop' },
    user: { id: b64(16), name: 'maria@exemplo.com', displayName: 'Maria' },
    challenge: b64(32),
    pubKeyCredParams: [{ type: 'public-key', alg: -7 }],
    authenticatorSelection: { residentKey: 'required', requireResidentKey: true, userVerification: 'required' },
    attestation: 'none',
    timeout: 60000,
  },
});

const requestOptions = () => ({
  options: { challenge: b64(32), rpId: 'localhost', userVerification: 'required', timeout: 60000 },
});

type State = { signedIn: boolean; bodies: Record<string, unknown>; csrf: (string | undefined)[] };

async function mockApi(page: Page, state: State) {
  await page.route('**/api/v1/version', (route) =>
    route.fulfill({ json: { api: '0.1.0', commit: 'abc1234', env: 'tst' } }),
  );
  await page.route('**/api/v1/me', (route) =>
    state.signedIn ? route.fulfill({ json: me }) : route.fulfill({ status: 401 }),
  );
  await page.route('**/api/v1/auth/providers', (route) => route.fulfill({ json: { providers: [] } }));
  await page.route('**/api/v1/auth/passkeys/**', (route) => {
    const path = new URL(route.request().url()).pathname;
    state.bodies[path] = route.request().postDataJSON();
    state.csrf.push(route.request().headers()['x-requested-with']);
    if (path.endsWith('/register/begin')) return route.fulfill({ json: creationOptions() });
    if (path.endsWith('/login/begin')) return route.fulfill({ json: requestOptions() });
    state.signedIn = true;
    if (path.endsWith('/register/finish')) {
      const created = { id: me.id, name: 'Passkey', created_at: '2026-10-15T12:00:00Z', backup_eligible: false };
      return route.fulfill({ status: 201, json: created });
    }
    return route.fulfill({ status: 204 });
  });
}

test('convite: cria a passkey no aparelho (virtual), manda ao servidor e entra logado', async ({ page }) => {
  const state: State = { signedIn: false, bodies: {}, csrf: [] };
  await addVirtualAuthenticator(page);
  await mockApi(page, state);
  await page.goto('/invite#segredo-do-convite');
  await expect(page).toHaveURL(/\/invite$/);
  await page.getByRole('button', { name: 'Criar minha passkey' }).click();
  await expect(page.getByText('Olá, Maria!')).toBeVisible();
  expect(state.bodies['/api/v1/auth/passkeys/register/begin']).toEqual({ invite: 'segredo-do-convite' });
  expect(state.csrf.length).toBeGreaterThan(0);
  expect(new Set(state.csrf)).toEqual(new Set(['XMLHttpRequest']));
  const finish = state.bodies['/api/v1/auth/passkeys/register/finish'] as {
    credential: { id: string; type: string; response: { attestationObject: string; clientDataJSON: string } };
  };
  expect(finish.credential.type).toBe('public-key');
  expect(finish.credential.response.attestationObject.length).toBeGreaterThan(20);
  expect(finish.credential.response.clientDataJSON.length).toBeGreaterThan(20);
});

test('entrar com passkey: o aparelho assina o desafio e o servidor recebe a asserção', async ({ page }) => {
  const state: State = { signedIn: false, bodies: {}, csrf: [] };
  await addVirtualAuthenticator(page);
  await mockApi(page, state);
  await page.goto('/invite#segredo');
  await page.getByRole('button', { name: 'Criar minha passkey' }).click();
  await expect(page.getByText('Olá, Maria!')).toBeVisible();

  await page.route('**/api/v1/auth/logout', (route) => {
    state.signedIn = false;
    return route.fulfill({ status: 204 });
  });
  await page.getByRole('button', { name: 'Sair' }).click();
  await expect(page.getByRole('button', { name: 'Entrar com passkey' })).toBeVisible();
  await page.getByRole('button', { name: 'Entrar com passkey' }).click();
  await expect(page.getByText('Olá, Maria!')).toBeVisible();
  const finish = state.bodies['/api/v1/auth/passkeys/login/finish'] as {
    credential: { response: { authenticatorData: string; signature: string; userHandle: string } };
  };
  expect(finish.credential.response.signature.length).toBeGreaterThan(20);
  expect(finish.credential.response.authenticatorData.length).toBeGreaterThan(20);
  expect(finish.credential.response.userHandle.length).toBeGreaterThan(10);
});

test('sem sessão: Entrar com passkey; com sessão: Olá, passkeys e Sair', async ({ page }) => {
  const state: State = { signedIn: false, bodies: {}, csrf: [] };
  await mockApi(page, state);
  await page.goto('/');
  await expect(page.getByRole('button', { name: 'Entrar com passkey' })).toBeVisible();
  state.signedIn = true;
  await page.reload();
  await expect(page.getByText('Olá, Maria!')).toBeVisible();
  await expect(page.getByTestId('version-footer')).toContainText('API v0.1.0 · tst');
  await page.getByRole('link', { name: 'Minhas passkeys' }).click();
  await expect(page).toHaveURL(/\/passkeys$/);
});

test('Sair manda o cabeçalho anti-CSRF e volta para Entrar com passkey', async ({ page }) => {
  const state: State = { signedIn: true, bodies: {}, csrf: [] };
  await mockApi(page, state);
  await page.route('**/api/v1/auth/logout', (route) => {
    expect(route.request().method()).toBe('POST');
    expect(route.request().headers()['x-requested-with']).toBe('XMLHttpRequest');
    state.signedIn = false;
    return route.fulfill({ status: 204 });
  });
  await page.goto('/');
  await page.getByRole('button', { name: 'Sair' }).click();
  await expect(page.getByRole('button', { name: 'Entrar com passkey' })).toBeVisible();
});

test('abrir /invite sem segredo mostra que o convite não vale', async ({ page }) => {
  await mockApi(page, { signedIn: false, bodies: {}, csrf: [] });
  await page.goto('/invite');
  await expect(page.getByRole('alert')).toContainText('não é válido');
});

test('login: cadastro e voltar, Google em breve e troca de tema', async ({ page }) => {
  await mockApi(page, { signedIn: false, bodies: {}, csrf: [] });
  await page.goto('/');
  await expect(page).toHaveURL(/\/login$/);
  await expect(page.getByRole('button', { name: /Continuar com Google/ })).toBeDisabled();
  await page.getByRole('link', { name: 'Criar conta' }).click();
  await expect(page).toHaveURL(/\/signup$/);
  await expect(page.getByRole('heading', { name: 'Criar conta' })).toBeVisible();
  await page.getByRole('link', { name: 'Entrar' }).click();
  await expect(page).toHaveURL(/\/login$/);

  const side = page.locator('section');
  await page.getByRole('button', { name: 'Usar tema escuro' }).click();
  await expect(side).toHaveCSS('background-color', 'rgb(18, 14, 36)');
  await page.getByRole('button', { name: 'Usar tema claro' }).click();
  await expect(side).toHaveCSS('background-color', 'rgb(255, 255, 255)');
});
