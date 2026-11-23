# Layout para desktop

> Arquivo: `docs/features/2026-11-23-layout-desktop.md` · Issue: #__ · Branch: `feat/layout-desktop` · Estado: PR

## Contexto
A web só tinha o layout mobile: conteúdo em coluna estreita (40–48rem) e o menu inferior fixo. Num monitor, o Início
virava uma faixa no meio da tela e o menu ficava longe do conteúdo. O dono do projeto quer manter a interface mobile
como está e ganhar um layout próprio para desktop. Outras frentes (seleção do estudo, meta semanal, ranking) vão
acrescentar telas ao menu, então o menu não pode ser um arquivo central que toda feature edita.

## Perguntas e respostas dos requisitos
| # | Pergunta | Recomendação | Resposta |
|---|---|---|---|
| 1 | O que muda no mobile? | Nada: o menu inferior e as telas continuam iguais abaixo de `bp.desktop` | Definido no pedido (23/11/2026) |
| 2 | Que navegação no desktop? | Menu lateral fixo com os mesmos itens; o inferior some | Definido no pedido (23/11/2026) |
| 3 | De onde vêm os itens do menu? | De cada `features/*/feature.ts` (registro por feature), lidos pelos dois menus | Definido no pedido (23/11/2026) |
| 4 | Como fica o Início no desktop? | Grade: atalhos e meta numa coluna, o painel de progresso (em duas colunas) ao lado | Definido no pedido (23/11/2026) |
| 5 | E a sessão de estudo? | Continua centrada em ~40rem, com atalhos de teclado 1–4 e Enter | Definido no pedido (23/11/2026) |

## Definições confirmadas
- Abaixo de 1100px (`bp.desktop`) o menu inferior continua exatamente como era.
- A partir de 1100px: menu lateral fixo (Perfil, Início, Estudar, Palavras, na mesma ordem) e sem menu inferior.
- Cada feature declara os próprios itens de menu em `feature.ts` (id, chave i18n, ícone, rota, ordem, rotas em que fica
  marcado). Uma feature nova (ex.: `leaderboard` com "Ranking") entra no menu só com arquivos dela.
- `AppPage` e `BrandScreen` aceitam `wide` para ir até 72rem no desktop; o Início usa isso e vira grade.
- A sessão de estudo fica centrada em 40rem; 1–4 escolhem a opção e Enter verifica/continua, sem atrapalhar campos de
  texto (o campo é dono das teclas).

## Decisões
| Tema | Decisão | Alternativa descartada / motivo |
|---|---|---|
| Registro | `core/nav/registry.ts` lê `features/*/feature.ts` com `import.meta.glob` (eager) e ordena por `order` | lista em `shared/`: arquivo central; `shared` importando `features`: inverte as camadas |
| Tipo do item | `shared/ui/nav/types.ts` (`NavItem`), usado pelas features e pelos menus | tipo em `core`: feature importaria `core` só para o tipo |
| Rótulo | chave com namespace (`common:nav.home`); os menus carregam os namespaces dos itens | chave só no `common`: Ranking teria de editar o `common.json` |
| Qual menu mostrar | `useMedia(desktopQuery)` (`matchMedia` + `useSyncExternalStore`) renderiza só um dos dois | os dois no DOM com `display` por CSS: dois "Menu principal" na árvore de acessibilidade e nos testes |
| Consulta de mídia em JS | `shared/ui/tokens/breakpoints.ts` com as mesmas larguras de `media.stylex.ts` (um teste garante que batem) | ler `bp.desktop` em tempo de execução: `defineConsts` é compilado pelo StyleX |
| Atalhos do estudo | já existiam (`useExerciseKeys`: A–D/1–4, Enter, 1–3 da dificuldade); só ganharam teste de campo de texto | reescrever: nada a mudar |
| Medidas | `size.content` (40rem), `size.page` (48rem), `size.wide` (72rem), `size.sideNav` (15rem) em `tokens` | literais espalhados |

## Checklist de progresso
- [x] Contrato: sem mudança de API
- [x] API: sem mudança
- [x] Web: registro de menu, menu lateral, larguras, grade do Início, tokens, testes (Vitest + Playwright)
- [x] Documentação: este arquivo
- [x] `npm run check` verde
- [x] PR aberta

## Verificação
- `npm run check` (web): lint, tipos, i18n, estrutura e Vitest verdes (resumo na PR).
- Playwright `e2e/layout/layout.spec.ts`: Início, Palavras e Estudar em 375px (menu inferior, uma coluna, sem rolagem
  horizontal) e 1280px (menu lateral, sem menu inferior, Início em grade).
