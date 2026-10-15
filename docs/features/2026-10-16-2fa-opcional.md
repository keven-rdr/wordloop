# 2FA opcional (ligado pelo admin) e criação do admin por SQL

> Arquivo: `docs/features/2026-10-16-2fa-opcional.md` · Issue: #__ · Branch: `feat/api-2fa-opcional` · Estado: BUILD

## Contexto
Hoje o admin **só entra** se o login provar 2 fatores. O dono não quer um 2FA imposto: quer que seja uma opção que o admin liga na própria conta (por ora só admin). Também quer criar o primeiro admin com um SQL rodado no Neon, sem SSH.

## Definições confirmadas
- 2FA **não é padrão**: vem desligado para todo mundo, inclusive admin.
- Quem pode ligar: **só administrador**, por enquanto. Os outros recebem 403.
- Admin continua com a sessão curta (8 h absolutas / 30 min de inatividade): isso não muda.
- Criar o admin por SQL no Neon, devolvendo o link do convite.

## Decisões
| Tema | Decisão | Alternativa descartada / motivo |
|---|---|---|
| Onde fica | coluna `app_user.mfa_required boolean NOT NULL DEFAULT false` (migração 00006, só aditiva) | tabela própria: um booleano por conta não precisa |
| Regra | `SessionOpener` recusa o login (403 no OIDC; 401 na passkey) só se `mfa_required` e a prova de 2 fatores faltar | manter "admin sempre exige": é o que o dono não quer |
| Senha | entrar só com senha **não** comprova 2 fatores (`MFA=false`); conta com o 2FA ligado recebe 403 e entra pela passkey ([ADR 0029](../adr/0029-senha-como-hash-argon2id-ao-lado-da-passkey.md)) | tratar a senha como 2º fator: ela é o 1º |
| Passkey | continua `userVerification: required`, então entrar com passkey já comprova 2 fatores | relaxar o UV: enfraquece todas as contas para uma opção |
| Ligar | `PUT /me/two-factor {enabled}` (204); só admin (403); **ligar exige sessão que já comprovou 2 fatores** (409), para o admin não se trancar fora ao ligar numa sessão fraca; desligar vale de qualquer sessão | permitir ligar de qualquer sessão: tranca fora quem entrou pelo Google |
| Mostrar | `GET /me` ganha `two_factor_enabled` | endpoint separado: um pedido a mais |
| Admin por SQL | `deploy/scripts/create-admin.sql`: cria o admin (ou reemite o convite de um admin pendente) e devolve o link; grava só o hash sha256 do segredo, igual ao `api create-admin` | gerar o segredo no navegador: o dono teria de calcular hash na mão |

## Checklist de progresso
- [x] Migração 00006, consulta `SetMFARequired`, `domain.User.MFARequired`
- [x] `SessionOpener` com a nova regra e testes (unidade e HTTP)
- [x] `accounts.TwoFactor`, `PUT /me/two-factor` e `two_factor_enabled` em `GET /me`
- [x] `deploy/scripts/create-admin.sql` validado em Postgres real (PGlite): cria, reemite, recusa quem já tem passkey ou é aluno
- [ ] Web: botão de ligar/desligar (vem junto com a tela de conta, depois do novo login)
- [x] `npm run check` verde
- [ ] PR aberta (com aprovação do usuário)

## Verificação
Automática: `go test ./...` (admin entra sem 2FA; com a opção ligada, só com prova; 403/409; aluno não liga). SQL: `create-admin.sql` rodado em PGlite com 9 conferências (hash igual ao `token.Hash`, convite antigo revogado, e-mail normalizado).
Manual em tst: rodar o SQL no Neon, abrir o link, criar a passkey, entrar, `GET /api/v1/me` mostra `two_factor_enabled=false`.
