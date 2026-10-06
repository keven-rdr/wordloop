import { defineConfig, devices } from '@playwright/test';

// Chromium, pt-BR e fuso fixo. A API e mockada por spec (e2e/<feature>/).
export default defineConfig({
  testDir: './e2e',
  reporter: process.env.CI ? [['list'], ['html', { open: 'never' }]] : 'list',
  use: { baseURL: 'http://localhost:4173', locale: 'pt-BR', timezoneId: 'America/Sao_Paulo' },
  projects: [{ name: 'chromium', use: { ...devices['Pixel 7'] } }],
  webServer: {
    command: 'npm run build && npx vite preview --port 4173 --strictPort',
    url: 'http://localhost:4173',
    reuseExistingServer: !process.env.CI,
    timeout: 120_000,
  },
});
