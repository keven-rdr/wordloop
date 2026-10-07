// Conventional Commits. O titulo da PR segue o mesmo padrao (workflow pr-title) porque o merge e por squash.
export default {
  extends: ['@commitlint/config-conventional'],
  rules: {
    'header-max-length': [2, 'always', 120],
    'body-max-line-length': [0],
    'footer-max-line-length': [0],
    'subject-case': [0], // portugues: nao forcar caixa
  },
};
