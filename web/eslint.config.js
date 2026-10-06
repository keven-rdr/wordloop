import stylex from '@stylexjs/eslint-plugin';
import tsParser from '@typescript-eslint/parser';

// typescript-eslint ainda exige a API do TypeScript 6: o pacote `typescript` e um alias de @typescript/typescript6,
// e o `tsc` 7 vem de @typescript/native (ver package.json). ESLint so para o que o Biome nao faz.
const features = ['auth', 'onboarding', 'study', 'dashboard', 'goals', 'settings', 'library', 'reading'];

export default [
  { ignores: ['src/api/generated/**', 'src/routeTree.gen.ts'] },
  {
    files: ['src/**/*.{ts,tsx}'],
    languageOptions: { parser: tsParser, parserOptions: { ecmaFeatures: { jsx: true } } },
    plugins: { '@stylexjs': stylex },
    rules: {
      '@stylexjs/valid-styles': 'error',
      '@stylexjs/no-unused': 'error',
      'max-lines': ['error', { max: 300, skipBlankLines: true, skipComments: true }],
      'max-lines-per-function': ['error', { max: 60, skipBlankLines: true, skipComments: true }],
      complexity: ['error', 10],
    },
  },
  // Uma feature nao importa outra feature (AGENTS.md): so shared/, core/ e api/generated/.
  ...features.map((name) => ({
    files: [`src/features/${name}/**/*.{ts,tsx}`],
    rules: {
      'no-restricted-imports': [
        'error',
        {
          patterns: features
            .filter((other) => other !== name)
            .map((other) => ({ group: [`**/features/${other}`, `**/features/${other}/**`], message: 'feature nao importa feature' })),
        },
      ],
    },
  })),
];
