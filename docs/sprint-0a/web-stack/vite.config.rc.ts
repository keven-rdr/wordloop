import react, { reactCompilerPreset } from '@vitejs/plugin-react';
import babel from '@rolldown/plugin-babel';
import stylex from '@stylexjs/unplugin';
import { defineConfig } from 'vitest/config';

export default defineConfig({
  plugins: [
    stylex.vite({ useCSSLayers: true, devMode: process.env.VITEST ? 'css-only' : 'full' }),
    react(),
    babel({ presets: [reactCompilerPreset()] }),
  ],
  test: { environment: 'jsdom', setupFiles: ['./src/test-setup.ts'], css: false },
});
