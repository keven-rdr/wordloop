# Entrar com e-mail e senha (admin primeiro)

> Arquivo: `docs/features/2026-10-16-senha-do-admin.md` · Issue: #__ · Branch: `feat/senha-do-admin` · Estado: BUILD

## Contexto
Depois de testar a tela, o dono disse que não quer entrar só com passkey: quer **e-mail e senha**, e perguntou como logar como admin. O ADR 0028 só previa passkey. A regra de fundo, porém, é "o app não pode saber a senha de ninguém": por isso a senha é guardada só como hash (ver [ADR 0029](../adr/0029-senha-como-hash-argon2id-ao-lado-da-passkey.md)).

## Definições confirmadas
- E-mail e senha no próprio app (resposta do dono: "E-mail e senha no app"); a passkey segue como opção.
- O admin é criado por SQL no Neon (`deploy/scripts/create-admin.sql`) e define a própria senha pelo link de convite, então a senha nunca passa por SQL, chat ou arquivo.
- 2FA continua sendo opção que o admin liga ([2FA opcional](2026-10-16-2fa-opcional.md)).

## Decisões
| Tema | Decisão | Alternativa descartada / motivo |
|---|---|---|
| Hash | Argon2id (`golang.org/x/crypto`), 19 MiB, 2 passadas, concorrência 2 | bcrypt: limite de 72 bytes e menos resistente a GPU |
| Definir a senha | o convite (15 min, uso único) abre a tela de definir senha; o convite é consumido na mesma transação que grava o hash | o admin informar a senha no SQL: a senha apareceria em histórico do editor |
| Login | `POST /auth/password/login`, 204 com cookie `httpOnly`; recusas sempre 401 igual; hash isca quando a conta não existe | 404/mensagens diferentes: revelariam quais e-mails existem |
| Bloqueio | a tentativa é contada **antes** da conferência, atômica no banco; a 5ª errada trava as seguintes por 30 s, dobrando até 15 min; tentativas antigas (> 15 min) somem; enquanto bloqueada, até a senha certa dá 401 | contar depois de conferir: uma rajada simultânea passaria do limite; bloqueio por IP só: um botnet ignora; bloqueio permanente: trava o dono fora |
| Trocar a senha pelo convite | grava o hash, consome o convite e **encerra todas as sessões antigas** na mesma transação | deixar as sessões: um cookie roubado sobreviveria à troca |
| Sem login por senha no ambiente | 503, como as passkeys | 400: confundiria com cabeçalho ausente |
| 2FA | senha sozinha não é 2 fatores: conta com 2FA ligado recebe 403 e entra pela passkey | aceitar a senha como 2º fator: ela é o 1º |
| Esqueci a senha (admin) | `deploy/scripts/reset-admin-access.sql`: novo convite, apaga as sessões e destrava a conta | recuperação por e-mail do admin: o e-mail viraria a chave mestra (decisão do ADR 0028) |
| Sessão | igual à da passkey (admin 8 h / 30 min ocioso), com `MFA=false` gravado | marcar `MFA=true`: seria mentira |
| Tela | login com e-mail e senha em primeiro lugar; passkey e Google logo abaixo; no convite a senha vem antes da passkey | esconder a passkey: o dono pode querer usá-la depois |

## Fora desta PR
Trocar a própria senha logado (`PUT /me/password`), cadastro de alunos por e-mail e recuperação por e-mail (PR 3, Brevo), Google e GitHub (PR 4), botão de ligar o 2FA na tela de conta.

## Checklist de progresso
- [x] Migração 00007 (`password_hash`, contadores e bloqueio), consultas sqlc
- [x] `platform/password` (Argon2id com limite de concorrência) e `identity/passwords` (login, definir senha, política)
- [x] Contrato: `POST /auth/password/login` e `POST /auth/password/set`; handlers e testes "estilo Postman"
- [x] SQL do Neon: `create-admin.sql` (considera senha) e `reset-admin-access.sql`, validados em Postgres
- [x] Web: login com e-mail e senha, definir a senha no convite, mostrar/ocultar senha, textos `pt-BR`/`en-US`
- [x] Testes: Go (unidade, HTTP, integração com Postgres real), Vitest e Playwright
- [x] Verificação real: Postgres 17 (Docker) + API + Vite + navegador: convite → senha → entrar → sair → entrar com e-mail e senha, senha errada, bloqueio, 2FA ligado (403)
- [x] `npm run check` verde
- [ ] PR aberta (com aprovação do usuário)

## Verificação
Automática: `go test ./...`, `go test -tags=integration ./...`, `npm run test`, `npm run test:e2e`.
Manual (dev): subir um Postgres, `api migrate`, rodar o SQL com `http://localhost:5173` no lugar do site, abrir o link, definir a senha, sair e entrar de novo.
Em tst: rodar `deploy/scripts/create-admin.sql` no Neon (trocando o e-mail e o nome), abrir o link e definir a senha.
