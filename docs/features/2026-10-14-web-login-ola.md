# Web: entrar, "Olá" e sair

> Arquivo: `docs/features/2026-10-14-web-login-ola.md` · Issue: #__ · Branch: `feat/web-login-ola` · Estado: VERIFY

## Contexto
Fecha o que falta do **walking skeleton do Sprint 1** (roadmap, seção 2): "login → tela 'Olá' autenticada com rodapé de versão → um E2E de fumaça". A API já faz o login (BFF, cookie `httpOnly`), `GET /me`, logout e sessões. O web ainda só mostra o nome do app e o rodapé.

## Definições confirmadas
Do planejamento (Fase 5 e roadmap), sem decisão nova:
- Entrar acontece **no provedor de identidade**; o web só leva a pessoa até `/api/v1/auth/login` (nada de formulário de senha no wordloop).
- Autenticação por **cookie**: o web nunca guarda token; só pergunta `GET /me`.
- Textos nos dois idiomas (`pt-BR`, `en-US`); alvo de toque ≥ 48 px; contraste AA; só tokens StyleX.
- O logout manda `X-Requested-With: XMLHttpRequest` (proteção CSRF do contrato).

## Decisões desta PR
| Tema | Decisão | Alternativa descartada / motivo |
|---|---|---|
| Roteamento | **sem roteador** ainda: uma tela que muda conforme a sessão (`/me`) | instalar o TanStack Router agora: só vale quando houver 2+ telas reais (Sprint 4) |
| Feature | `src/features/auth/` (nome já previsto no ESLint e na Fase 5); textos no namespace `auth` | misturar no `common`: o namespace carrega sob demanda |
| Entrar | link nativo `<a href="/api/v1/auth/login?return_to=/">` com cara de botão | `fetch`: o login é redirecionamento de página inteira |
| Estados | carregando (nada de botão, evita piscar) · anônimo (Entrar) · logado ("Olá, {nome}!" e Sair) · erro no `/me` conta como anônimo | tela de erro: `/me` dá 401 para quem não entrou, é o caminho normal |
| Sair | `POST /auth/logout` com o cabeçalho; depois refaz o `/me` (que passa a dar 401). Falha mostra aviso com `role="alert"` | recarregar a página |
| Papel (`role`) | **não** usado para esconder nada; a API confere o papel a cada pedido | esconder botões de admin: o front não é barreira de segurança |
| Teste | Vitest por componente + Playwright com a API mockada (3 estados) | contra o tst: depende do provedor de login (ver abaixo) |

## Fora desta PR
- **Provedor de login do tst** (`[DECISÃO SUA]`): sem `OIDC_*` o botão Entrar leva a um 503. A tela já funciona com o provedor ligado; o teste de fumaça contra o tst entra quando existir.
- Roteador, `/welcome`, onboarding, menu: Sprint 4 em diante.
- Lista de sessões ("meus aparelhos") na interface: o endpoint existe; a tela entra com as configurações (Sprint 5).

## Checklist de progresso
- [x] Contrato já pronto (`getMe`, `logout`); cliente regenerado (`npm run gen`)
- [x] `features/auth`: hook de sessão, painel (Entrar / Olá + Sair), namespace `auth` em `pt-BR` e `en-US`
- [x] Testes: Vitest (3 estados + logout) e Playwright (`e2e/auth/`)
- [x] `npm run check` verde
- [ ] PR aberta (com aprovação do usuário)

## Verificação
Automática: `npm run test` e `npm run test:e2e`. Manual (dev): subir o Compose, API com `OIDC_*` do Keycloak e `npm run dev:web`; abrir `/`, clicar em Entrar, voltar logado com "Olá" e sair.
