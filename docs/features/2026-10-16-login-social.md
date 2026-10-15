# Login com Google e GitHub (vinculado por e-mail confirmado)

> Arquivo: `docs/features/2026-10-16-login-social.md` · Issue: #__ · Branch: `feat/login-social` · Estado: BUILD · PR 4 de 4 do plano de [contas próprias](2026-10-14-contas-proprias.md)

## Contexto
A tela de entrar tinha os botões do Google e do GitHub desligados ("em breve"). O dono quer entrar e criar conta com eles, **sem amarrar o app a nenhum provedor**: a conta é nossa e o provedor é só uma forma de provar o e-mail.

## Definições confirmadas
- Google **e** GitHub entram agora (resposta do dono em 15/10/2026).
- O login social **cria a conta se o e-mail ainda não existe e vincula se existe** (mesma resposta).
- O vínculo é só por **e-mail confirmado pelo provedor**.

## Decisões
| Tema | Decisão | Alternativa descartada / motivo |
|---|---|---|
| Quem é a pessoa | `auth_identity (provider, subject)` primeiro; sem vínculo, o e-mail confirmado acha a conta (e cria o vínculo); sem conta, cria uma já com o e-mail confirmado | vincular por e-mail não confirmado: quem criar uma conta GitHub com o e-mail de outra pessoa entraria na conta dela |
| E-mail não confirmado | recusa (`/login?error=email`), nada é criado nem vinculado | criar conta sem e-mail: sem como recuperar nem vincular depois |
| Cadastro pendente | se o e-mail tinha um cadastro por e-mail ainda não confirmado, o provedor o confirma; ele continua sem senha | deixar pendente: a pessoa ficaria com duas contas |
| Admin | é alcançado pelo e-mail confirmado como qualquer conta, mantendo o papel; o 2FA ligado na conta recusa o login social (`/login?error=two_factor`) | proibir o admin de usar provedor: o admin escolhe ligar o 2FA |
| Vários provedores | `app.Providers` por nome (`google`, `github`, `oidc`); o nome escolhido viaja **selado** no cookie do fluxo, então o callback não pode ser trocado de provedor | um callback por provedor: duas URLs para registrar |
| Google | OIDC (`accounts.google.com`) com PKCE e nonce, escopos `openid profile email` | biblioteca própria |
| GitHub | OAuth2 com PKCE (`read:user user:email`); o e-mail vem de `GET /user/emails`: o primário **confirmado**, senão qualquer confirmado | usar o e-mail público do perfil: não é confirmado |
| Falha no retorno | em vez de uma página em branco (400/403), o callback redireciona para `/login?error=failed\|two_factor\|email` e a tela mostra o motivo; o código de erro é uma lista fechada (valor desconhecido na URL é ignorado) | repassar o erro do provedor na URL: vazaria texto de terceiros |
| Quais botões | `GET /auth/providers` lista o que está ligado; a tela só habilita esses e mostra "em breve" nos outros | decidir por variável do build: exigiria rebuild para ligar |
| Sessão | `MFA=false` (o provedor não prova 2 fatores) | confiar em `amr` do Google: ele não informa |

## Limitações conhecidas
- Mudar o e-mail na conta do Google/GitHub **não** muda o e-mail da conta no wordloop (o vínculo é pelo `subject`, que não muda).
- Conta apagada segura o e-mail: o login social falha com erro (ainda não existe exclusão de conta; a regra da LGPD virá no Sprint 6).
- Sem tela para desvincular provedores.

## Checklist de progresso
- [x] `app.Providers`, provedor no fluxo selado, `ErrUnknownProvider` e `ErrEmailNotVerified`
- [x] Adaptador do GitHub (OAuth2 + PKCE) e claims de e-mail no OIDC (Google)
- [x] Regra de vínculo/criação por e-mail confirmado no Postgres
- [x] Contrato: `provider` em `/auth/login`, `GET /auth/providers`, falhas do callback por redirecionamento
- [x] Web: botões do Google e do GitHub habilitados pela lista do servidor, mensagens de `/login?error=`
- [x] Deploy: secrets `GOOGLE_*`, `GH_OAUTH_*` e `SESSION_KEY`
- [x] Testes: Go (unidade, HTTP, Postgres real), Vitest e Playwright
- [x] `npm run check` verde
- [ ] PR aberta (com aprovação do usuário)

## O que o dono precisa fazer para ligar em tst
Criar o cliente OAuth no Google e o OAuth App no GitHub e cadastrar os secrets: passo a passo em [`docs/ci-cd.md`](../ci-cd.md#login-social-em-tst).

## Verificação
Automática: `go test ./...`, `go test -tags=integration ./...`, `npm run test`, `npm run test:e2e`.
Manual (dev): com o Keycloak do Compose, `/api/v1/auth/login?provider=oidc`. Em tst, depois dos secrets: `GET /api/v1/auth/providers` e um login de verdade com cada provedor.
