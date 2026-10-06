# web/ — regras locais (React + TypeScript + StyleX)

Vale junto com o `AGENTS.md` da raiz (que vence em conflito).

## Estrutura
- `src/features/<nome>/` autocontida: `feature.ts` (id, menu, namespaces), `components/`, `hooks/`, `exercises/` (só `study`). Feature **não importa outra feature**; usa só `shared/`, `core/` e `api/generated/`.
- **Criar uma feature não edita arquivo central:** rota = arquivo novo em `src/routes/` (poucas linhas, importa a feature); menu = `feature.ts`; i18n = `src/locales/{pt-BR,en-US}/<ns>.json` novo.
- Proibido: `services/`, `queries/`, `schemas/`, `utils/` globais; componente/arquivo > 300 linhas.

## Dados
- Use os **hooks gerados** em `src/api/generated` (TanStack Query). Não escreva fetch, tipos ou schemas à mão para a API.
- Formulários: React Hook Form + Valibot (schemas gerados quando existirem). Erro de API (`code`, `errors[]`) → `useProblemHandler`.
- Estado de UI local em Zustand; estado do servidor **só** no TanStack Query.

## i18n
- `const { t } = useTranslation("study")`; textos comuns: `const { t: tCommon } = useTranslation("common")`.
- **Nunca texto literal em JSX**, nem em `title`, `placeholder`, `aria-label`, `alt`. Chave nova → nos dois idiomas, depois `npm run gen:i18n`.

## Estilo (StyleX)
- Só `stylex.create` com **tokens** (`shared/ui/tokens/*.stylex.ts`): cores, espaço, raio, tipografia, movimento. Breakpoints por `bp.*` (mobile-first); `prefers-reduced-motion` respeitado.
- Alvo de toque ≥ 48 px; contraste WCAG 2.2 AA; sem `style={{}}` (exceto valor dinâmico via função do `stylex.create`).
- Versões `@stylexjs/*` **exatas** e sempre juntas.
- **Condições dentro do valor da propriedade**, nunca no nível do objeto: `transitionDuration: { default: '120ms', [bp.reduceMotion]: '0s' }`, `boxShadow: { default: …, ':active': … }` (o StyleX rejeita `[bp.x]: { … }` solto).
- `tsc` é o TypeScript 7 (`@typescript/native`); `typescript` é um alias do TS 6 só para ferramentas (typescript-eslint, hey-api). Não troque.

## Regras de código (Sonar, aplicadas nas linhas alteradas)
`??` em vez de `||` para valor nulável (S6606) · sem ternário aninhado (S3358) · sem `as` desnecessário (S4325) · props `Readonly` (S6759) · elemento nativo em vez de role ARIA (S6819) · sem template literal aninhado (S4624) · classes de regex concisas (S6353).

## Testes
- Vitest + Testing Library por componente/hook; Playwright por feature (specs com mocks de API em `e2e/<feature>/`), mais poucos *smoke tests* contra a API real em tst.
- Comandos: `npm run test` · `npm run test:e2e` · `npm run check`.
