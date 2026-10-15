import babel from '@rolldown/plugin-babel';
import stylex from '@stylexjs/unplugin';
import { tanstackRouter } from '@tanstack/router-plugin/vite';
import react, { reactCompilerPreset } from '@vitejs/plugin-react';
import { defineConfig } from 'vitest/config';

const version = process.env.VERSION ?? 'dev';
const commit = (process.env.COMMIT ?? 'local').slice(0, 7);
const apiTarget = process.env.API_PROXY_TARGET || 'http://localhost:8080';
const compiler = process.env.VITEST ? [] : [babel({ presets: [reactCompilerPreset()] })]; // React Compiler 1.0 (validado com StyleX no Sprint 0a); fora dos testes, para a cobertura medir o nosso código e não o cache de memoização gerado

export default defineConfig({
  plugins: [
    tanstackRouter({ target: 'react' }), // ANTES do plugin do React: gera src/routeTree.gen.ts a partir de src/routes
    stylex.vite({ useCSSLayers: true, devMode: process.env.VITEST ? 'css-only' : 'full' }), // StyleX ANTES do plugin do React
    react(),
    ...compiler,
  ],
  define: {
    __APP_VERSION__: JSON.stringify(version),
    __APP_COMMIT__: JSON.stringify(commit),
  },
  server: { proxy: { '/api': apiTarget } },
  test: {
    environment: 'jsdom',
    setupFiles: ['./src/test-setup.ts'],
    css: false,
    testTimeout: 15_000,
    exclude: ['e2e/**', 'node_modules/**'],
    coverage: {
      provider: 'v8',
      include: ['src/**'],
      exclude: [
        'src/api/generated/**',
        'src/**/*.test.{ts,tsx}',
        'src/test-setup.ts',
        'src/test-support.tsx',
        'src/main.tsx', // bootstrap: coberto pelo E2E; o provedor v8 do Vitest 5 não consegue analisar TSX não importado
        'src/vite-env.d.ts',
        'src/routeTree.gen.ts',
      ],
    },
  },
});
