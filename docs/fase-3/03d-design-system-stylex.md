# Fase 3 — Design system em StyleX

> Versões confirmadas em 06/10/2026 no registry do npm: `@stylexjs/stylex`, `@stylexjs/unplugin`, `@stylexjs/eslint-plugin` **0.19.1** (pré-1.0; release de 15/09/2026 no [GitHub](https://github.com/facebook/stylex/releases)); `@base-ui/react` **1.8.0**; `react-aria-components` **1.21.1**; Vite **8.3.3**; Vitest **5.0.3**; TypeScript **7.0.2**.
> Os trechos abaixo ilustram a API e **não foram compilados**; o spike do Sprint 0 valida tudo (seção 5).

## 1. Tokens (`defineVars`) e temas (`createTheme`)

Variáveis só existem em arquivos `.stylex.ts` com *exports* nomeados. Nenhuma cor, espaço ou raio crus fora daqui.

```ts
// shared/ui/tokens/colors.stylex.ts
import * as stylex from '@stylexjs/stylex';

export const color = stylex.defineVars({
  bg:        { default: '#FFFFFF', '@media (prefers-color-scheme: dark)': '#101820' },
  surface:   { default: '#F4F6F8', '@media (prefers-color-scheme: dark)': '#1A2530' },
  text:      { default: '#16202A', '@media (prefers-color-scheme: dark)': '#EAF0F5' },
  primary:   { default: '#0E9F8E', '@media (prefers-color-scheme: dark)': '#2CC4B2' },
  primaryEdge: { default: '#0A7A6D', '@media (prefers-color-scheme: dark)': '#1B9C8C' }, // "sombra 3D" do botão
  success: '#2E9E4F', danger: '#D64545', warning: '#E0A100',
});

// shared/ui/tokens/themes.stylex.ts — escolha EXPLÍCITA do usuário (a media query cobre só "sistema")
export const lightTheme = stylex.createTheme(color, { bg: '#FFFFFF', surface: '#F4F6F8', text: '#16202A', primary: '#0E9F8E', primaryEdge: '#0A7A6D' });
export const darkTheme  = stylex.createTheme(color, { bg: '#101820', surface: '#1A2530', text: '#EAF0F5', primary: '#2CC4B2', primaryEdge: '#1B9C8C' });
```
Na raiz: `system` → nenhuma classe de tema (vale a `@media`); `light`/`dark` → `stylex.props(lightTheme)` / `stylex.props(darkTheme)` no `<html>`. A paleta é **nossa** (teal), sem o verde, a fonte, o mascote nem os assets do Duolingo; os valores acima são marcadores a ajustar no Sprint 1 com contraste WCAG 2.2 AA verificado.

Demais famílias de tokens: `space` (4, 8, 12, 16, 24, 32), `radius`, `font` (tamanhos, pesos, famílias do sistema para começar), `shadow`, `motion` (durações/curvas), `z`.

## 2. Breakpoints e constantes (`defineConsts`)

```ts
// shared/ui/tokens/media.stylex.ts
export const bp = stylex.defineConsts({
  tablet:  '@media (min-width: 768px)',
  desktop: '@media (min-width: 1100px)',
  reduceMotion: '@media (prefers-reduced-motion: reduce)',
});
// uso:  width: { default: '100%', [bp.desktop]: 480 }
```
`defineConsts` valores são **inline** em tempo de build (sem variável CSS), por isso servem a media queries e *z-index*. Limitação documentada: sem `enableMediaQueryOrder` em constantes ([docs](https://stylexjs.com/docs/api/javascript/defineConsts)). Mobile-first: o estilo base é o de celular; `bp.*` só amplia.

## 3. Dinâmico, animação, botão "3D"

```ts
const styles = stylex.create({
  bar:  { height: 12, borderRadius: radius.full, backgroundColor: color.surface },
  fill: (pct: number) => ({ width: `${pct}%`, backgroundColor: color.primary, transitionDuration: motion.base }), // barra de progresso
  heat: (level: 0|1|2|3|4) => ({ opacity: [0.12, 0.35, 0.55, 0.8, 1][level] }),                                    // mapa de calor
  button: { backgroundColor: color.primary, boxShadow: `0 4px 0 ${color.primaryEdge}`, minHeight: 48,
            ':active': { transform: 'translateY(4px)', boxShadow: `0 0 0 ${color.primaryEdge}` },
            [bp.reduceMotion]: { transitionDuration: '0s' } },
});
const pop = stylex.keyframes({ from: { transform: 'scale(0.9)' }, to: { transform: 'scale(1)' } });
```
Funções em `stylex.create` geram estilo dinâmico por variável CSS inline (sem classe nova por valor). Alvo de toque ≥ 48 px.

## 4. Primitivas do MVP

| Primitiva | Base | Observação |
|---|---|---|
| Button ("3D") | própria | variantes: primário, secundário, perigo; estados *pressed* e *disabled* |
| Card | própria | |
| ProgressBar | Base UI `Progress` | meta do dia e da sessão |
| Dialog / Sheet | Base UI `Dialog` | sheet inferior no celular |
| Tabs | Base UI `Tabs` | |
| Toast | Base UI `Toast` | feedback certo/errado não usa toast (fica inline no exercício) |
| Input | própria + `<input>` nativo | digitar resposta; `autoCapitalize=off`, `spellCheck=false` |
| Switch | Base UI `Switch` | configurações |
| Heatmap | **SVG próprio** | 7×N células, intensidade por token; sem biblioteca de gráficos |
| ChoiceButton, FeedbackBanner, DifficultyPicker | próprias | núcleo da tela de exercício |

**Por que não Astryx já:** [facebook/astryx](https://github.com/facebook/astryx) existe (MIT, ~13,5 mil estrelas, ativo hoje; pacotes `@astryxdesign/core`, `/build`, `/cli`, `/theme-*`; beta; React 19+; vem com CSS pré-compilado). É útil como **referência de padrões e como fonte de ideias**, mas é beta, tem identidade visual própria e traria 150 componentes que não usaremos. **Base UI** (estável, 1.8.0) dá comportamento e acessibilidade sem estilo, que é o que o StyleX pede. Reavaliar Astryx quando sair do beta.
**Base UI × React Aria Components:** escolho Base UI para começar; React Aria é a alternativa se precisarmos de recursos de internacionalização/acessibilidade que o Base UI não cubra `[VERIFICAR]` no spike.

## 5. Riscos do StyleX e mitigação

| Risco | Mitigação |
|---|---|
| Pré-1.0: quebras entre versões | versões **exatas** (sem `^`) dos três pacotes `@stylexjs/*`, atualizados juntos; PR de Renovate agrupada e E2E visual obrigatório |
| Ecossistema pequeno (sem "shadcn") | primitivas Base UI + 10 componentes próprios; Astryx como referência |
| Integração com Vite 8 | `stylex.vite()` **antes** de `react()` ([README do unplugin](https://github.com/facebook/stylex/tree/main/packages/%40stylexjs/unplugin)); spike no Sprint 0 com build, HMR e produção |
| Vitest 5 e StyleX | rodar o mesmo `vite.config` nos testes; se houver atrito, mockar estilos nos testes de unidade e cobrir visual no Playwright `[VERIFICAR]` |
| React Compiler 1.0 + StyleX | no `plugin-react` 6 o Babel é opt-in (`@rolldown/plugin-babel`); o StyleX usa seu próprio Babel. **Não habilitar o Compiler no MVP**; testar no spike e decidir `[VERIFICAR]` |
| Lint | Biome para o geral + **ESLint só** com `@stylexjs/eslint-plugin` (`valid-styles`, `no-unused`) e as regras de tamanho (`max-lines`, `complexity`) |
| Depuração de CSS atômico | `dev: true` gera nomes legíveis; `useCSSLayers` evita briga de especificidade |
| Projeto abandonado | os tokens viram variáveis CSS comuns; migrar para CSS Modules seria mecânico (custo alto, não zero) |
