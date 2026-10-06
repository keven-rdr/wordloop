import stylex from '@stylexjs/eslint-plugin';
import tsParser from '@typescript-eslint/parser';

// typescript-eslint ainda exige a API do TypeScript 6: o pacote `typescript` é um alias de @typescript/typescript6,
// e o `tsc` 7 vem de @typescript/native (ver package.json).
export default [
  {
    files: ['src/**/*.{ts,tsx}'],
    languageOptions: { parser: tsParser, parserOptions: { ecmaFeatures: { jsx: true } } },
    plugins: { '@stylexjs': stylex },
    rules: {
      '@stylexjs/valid-styles': 'error',
      '@stylexjs/no-unused': 'error',
      'max-lines': ['error', { max: 300, skipBlankLines: true, skipComments: true }],
      complexity: ['error', 10],
    },
  },
];
