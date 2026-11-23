// A mesma largura do desktop de `media.stylex.ts`, como consulta de `matchMedia` (o StyleX compila o `defineConsts`, então
// o JS não o lê em tempo de execução). `breakpoints.test.ts` garante que batem.
export const desktopQuery = '(min-width: 1100px)';
