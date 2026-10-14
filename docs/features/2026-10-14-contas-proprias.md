# Contas próprias: e-mail, vínculo de provedores e admin criado pelo dono

> Arquivo: `docs/features/2026-10-14-contas-proprias.md` · Issue: #__ · Branch: `feat/contas-proprias` (PR 1 de 4) · Estado: VERIFY

## Contexto
O dono não quer que o app conheça senha de ninguém, nem ficar preso a um provedor externo (Keycloak, Google, GitHub), nem um servidor de identidade pesado. O caminho aprovado: **cadastro no próprio app (Neon)**, **entrada por passkeys** (o app guarda só chaves públicas) e **Google/GitHub depois, vinculados**. Isso **substitui o [ADR 0010](../adr/0010-um-keycloak-com-dois-realms-tst-e-prd.md)** (Keycloak) para identidade; ver o [ADR 0028](../adr/0028-contas-proprias-sem-senha-passkeys-depois-google-github.md).

Hoje o `app_user` identifica a pessoa por `(idp_issuer, idp_subject)` e **não guarda e-mail**: trocar de provedor criaria usuários novos. Esta PR é a **fundação**: nenhum login novo ainda.

## Plano em 4 PRs
| PR | Entrega |
|---|---|
| **1 (esta)** | e-mail no perfil; tabela `auth_identity` (vínculo de provedores, sem amarrar o cadastro a nenhum); tabela `auth_token` (tokens de convite, só o hash); comando `api create-admin` que cria o admin e emite um convite de uso único |
| 2 | passkeys: registrar (com o convite) e entrar; gestão de passkeys; roteador no web; limite de tentativas |
| 3 | cadastro por e-mail e recuperação (Brevo); admin só recupera por comando no servidor |
| 4 | Google e GitHub vinculados pelo e-mail verificado |

## Definições confirmadas (pelo dono)
- O app **nunca guarda senha**. Entrar = passkey (PR 2), depois provedores sociais (PR 4).
- Primeiro se **cadastra o admin no app**; só depois entram Google e GitHub.
- E-mail por Brevo gratuito desde o início (PR 3).

## Decisões desta PR
| Tema | Decisão | Motivo |
|---|---|---|
| Identidade da pessoa | o **`app_user.id` é nosso**; provedores são só vínculos em `auth_identity (provider, subject)` | trocar ou somar provedores nunca recria usuários |
| Migração | **aditiva**: o login OIDC atual continua; `auth_identity` é preenchida a partir de `idp_issuer/idp_subject` e novos usuários gravam os dois (a versão antiga da API segue funcionando durante o deploy). Se uma linha antiga ainda não tem vínculo, o login a vincula na hora | expand/contract |
| E-mail | `app_user.email` único sem diferenciar maiúsculas (`lower(email)`); `email_verified_at` guarda quando foi verificado (o admin criado pelo dono nasce verificado: o dono atesta) | PR 4 liga o login social pelo e-mail verificado |
| Convite | token aleatório de 256 bits, **só o hash** (SHA-256) no banco, validade de 15 minutos, uso único (consumo na PR 2); link com o token no **fragmento** (`/invite#<token>`), que não vai ao servidor nem a logs | evita vazar o token por log ou `Referer` |
| `create-admin` | cria o usuário `admin` (e-mail normalizado em minúsculas) e imprime o link na saída padrão; se o e-mail já é de um admin **sem nenhuma credencial**, reemite o convite (os anteriores deixam de valer); qualquer outro caso é recusado | não permite tomar conta alheia |
| Onde roda | no servidor, pelo dono: `docker compose … run --rm api create-admin --email … --name …` (a chave de deploy é de comando forçado e não serve) | só quem tem acesso ao servidor cria admin |

## Fora desta PR
Login por passkey, tela de convite, roteador, limite de tentativas (PR 2); e-mail e cadastro aberto (PR 3); Google e GitHub (PR 4); remover `idp_issuer/idp_subject` (`contract`, depois da PR 4).

## Checklist de progresso
- [x] Migração 00004 (aditiva) e consultas `sqlc`
- [x] `Users.Upsert` por `auth_identity` (com a regra do vínculo antigo) e testes de integração
- [x] `internal/identity/accounts`: `CreateAdmin` (e-mail, convite, reemissão) com testes
- [x] Comando `api create-admin` e `APP_PUBLIC_URL` (`deploy/tst/api.env`)
- [x] ADR 0028 (e ADR 0010 marcado como substituído)
- [x] `npm run check` verde
- [ ] PR aberta (com aprovação do usuário)

## Verificação
Feito: `npm run check` verde; testes de unidade do convite (hash, reemissão, não tomar conta alheia, e-mail e URL inválidos); migração e consultas conferidas num Postgres real (PGlite) sobre linhas existentes, inclusive a corrida de dois primeiros logins (a segunda criação dá `23505` e não deixa linha órfã). Os testes de integração (Postgres e Keycloak reais) compilam e rodam no CI.

Automática: `go test ./...`; integração no CI (Postgres real); migração conferida em PGlite, inclusive sobre linhas existentes. Manual em dev: `api migrate`, `api create-admin --email voce@exemplo.com --name Voce` imprime o link; rodar de novo reemite; com o e-mail de outro papel, recusa.
