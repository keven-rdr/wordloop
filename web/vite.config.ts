import stylex from '@stylexjs/unplugin';
import react from '@vitejs/plugin-react';
import { defineConfig } from 'vitest/config';

// VERSION e COMMIT vem do CI (git describe / sha curto); localmente ficam "dev".
const version = process.env.VERSION ?? 'dev';
const commit = (process.env.COMMIT ?? 'local').slice(0, 7);

export default defineConfig({
  plugins: [stylex.vite({ useCSSLayers: true, devMode: process.env.VITEST ? 'css-only' : 'full' }), react()], // StyleX ANTES do plugin do React
  define: {
    __APP_VERSION__: JSON.stringify(version),
    __APP_COMMIT__: JSON.stringify(commit),
  },
  server: { proxy: { '/api': 'http://localhost:8080' } },
  test: {
    environment: 'jsdom',
    setupFiles: ['./src/test-setup.ts'],
    css: false,
    exclude: ['e2e/**', 'node_modules/**'],
  },
});
