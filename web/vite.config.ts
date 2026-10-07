import babel from '@rolldown/plugin-babel';
import stylex from '@stylexjs/unplugin';
import react, { reactCompilerPreset } from '@vitejs/plugin-react';
import { defineConfig } from 'vitest/config';

const version = process.env.VERSION ?? 'dev';
const commit = (process.env.COMMIT ?? 'local').slice(0, 7);

export default defineConfig({
  plugins: [
    stylex.vite({ useCSSLayers: true, devMode: process.env.VITEST ? 'css-only' : 'full' }), // StyleX ANTES do plugin do React
    react(),
    babel({ presets: [reactCompilerPreset()] }), // React Compiler 1.0 (validado com StyleX no Sprint 0a)
  ],
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
    coverage: {
      provider: 'v8',
      include: ['src/**'],
      exclude: [
        'src/api/generated/**',
        'src/**/*.test.{ts,tsx}',
        'src/test-setup.ts',
        'src/main.tsx', // bootstrap: coberto pelo E2E; o provedor v8 do Vitest 5 não consegue analisar TSX não importado
        'src/vite-env.d.ts',
        'src/routeTree.gen.ts',
      ],
    },
  },
});
