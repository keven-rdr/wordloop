import react from '@vitejs/plugin-react';
import stylex from '@stylexjs/unplugin';
import { defineConfig } from 'vitest/config';

export default defineConfig({
  plugins: [stylex.vite({ useCSSLayers: true, devMode: process.env.VITEST ? 'css-only' : 'full' }), react()], // StyleX ANTES do plugin do React
  test: {
    environment: 'jsdom',
    setupFiles: ['./src/test-setup.ts'],
    css: false,
  },
});
