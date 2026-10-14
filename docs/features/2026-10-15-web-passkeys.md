# Web: entrar com passkey, aceitar convite e "minhas passkeys"

> Arquivo: `docs/features/2026-10-15-web-passkeys.md` · Issue: #__ · Branch: `feat/web-passkeys` (PR 2b de 4) · Estado: BUILD

## Contexto
A [PR 2a](2026-10-15-passkeys-api.md) entregou a API das passkeys. Esta PR dá a **cara** a ela: o botão "Entrar com passkey", a tela que o link de convite abre (para o admin criar a primeira passkey) e a lista "minhas passkeys". Com isso fecha o Sprint 1: convite → passkey → entrar → "Olá" → sair, no tst.

## Definições confirmadas
- Entrar é por passkey; o web nunca vê senha nem token (cookie `httpOnly`). Ver [ADR 0028](../adr/0028-contas-proprias-sem-senha-passkeys-depois-google-github.md).
- O convite vem no **fragmento** do link (`/invite#<segredo>`): o navegador não o envia ao servidor.
- Textos em `pt-BR` e `en-US`; alvo de toque ≥ 48 px; só tokens StyleX.

## Decisões
| Tema | Decisão | Alternativa descartada / motivo |
|---|---|---|
| Roteador | **TanStack Router** por arquivo (`src/routes`), como no ADR 0014: agora há 3 telas (`/`, `/invite`, `/passkeys`). A árvore `routeTree.gen.ts` **não é versionada**: `npm run gen` a gera (CLI `tsr generate`) e o plugin do Vite a mantém em dev e no build | rotas escritas à mão: contraria o ADR e cresce mal |
| API do navegador | `PublicKeyCredential.parseCreationOptionsFromJSON` / `parseRequestOptionsFromJSON` e `credential.toJSON()` (sem conversão manual de base64url). Navegador sem isso mostra aviso claro | biblioteca extra (`@simplewebauthn/browser`): mais uma dependência para o que o navegador já faz |
| Segredo do convite | lido do fragmento e **apagado da barra de endereço** logo ao abrir (`history.replaceState`); fica só na memória da tela | deixar na URL: vaza em captura de tela e no histórico |
| Cancelar | o usuário cancelar o pedido do aparelho (`NotAllowedError`) mostra "Cancelado", sem alarme; outros erros mostram aviso e ficam no `console` só como nome do erro | tratar tudo como falha |
| Início | o botão **Entrar** deixa de levar ao OIDC (que nem está ligado em tst): vira "Entrar com passkey". Google e GitHub voltam na PR 4 | manter dois botões, um deles quebrado |
| Apagar passkey | a API recusa apagar a última (409); a tela mostra o motivo em vez de esconder o botão | esconder o botão: o usuário não entende por quê |

## Fora desta PR
Cadastro por e-mail e recuperação (PR 3); Google e GitHub (PR 4); renomear passkey; dar a cada tela um layout de menu (Sprint 4).

## Checklist de progresso
- [x] Roteador, geração de rotas (`npm run gen`, CI, Docker) e `App` com `RouterProvider`
- [x] `features/auth`: chamadas WebAuthn, hooks, `AuthPanel` com passkey, `InvitePage`, `PasskeysPage`
- [x] Textos `pt-BR` e `en-US` (namespace `auth`)
- [x] Testes: Vitest (mock do navegador) e Playwright com **autenticador virtual** (`e2e/auth/`)
- [x] Verificação real: API Go + Postgres + navegador com autenticador virtual (convite → passkey → entrar → sair → entrar → listar)
- [x] `npm run check` verde
- [ ] PR aberta (com aprovação do usuário)

## Verificação
Automática: `npm run test` e `npm run test:e2e`. Manual (dev): `npm run dev:api` com `WEBAUTHN_RP_ID=localhost` e `WEBAUTHN_ORIGIN=http://localhost:5173`, `api create-admin`, abrir o link impresso no navegador de `npm run dev:web` (o DevTools tem um autenticador virtual em *WebAuthn*).
