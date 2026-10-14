# ADR 0028 — Contas próprias sem senha: passkeys, depois Google e GitHub vinculados

**Status:** aceita (2026-10-14); o ponto "nenhuma senha" foi revisado pelo [ADR 0029](0029-senha-como-hash-argon2id-ao-lado-da-passkey.md) (senha como hash Argon2id, ao lado da passkey). **Substitui o [ADR 0010](0010-um-keycloak-com-dois-realms-tst-e-prd.md)** (Keycloak) e revisa o [ADR 0011](0011-api-como-bff-oidc-com-cookie-httponly.md) apenas no que diz respeito a *quem emite a identidade*: a API continua sendo o BFF com cookie `httpOnly` e sessão no servidor.

## Contexto
O dono não quer que o app conheça a senha de ninguém, nem ficar preso a um provedor externo, nem manter um servidor de identidade pesado (o Keycloak não cabe na VM de 1 GB de tst e é mais um serviço). Ele quer **cadastrar primeiro o admin no próprio app** (banco Neon) e só depois ligar Google e GitHub. O `app_user` identificava a pessoa por `(idp_issuer, idp_subject)` e não guardava e-mail: trocar de provedor criaria usuários novos.

## Decisão
- O **`app_user.id` é nosso**. Provedores são vínculos em `auth_identity (provider, subject)`; o e-mail (único, sem diferenciar maiúsculas) é do perfil.
- **Nenhuma senha, em lugar nenhum.** Entrar é por **passkey** (WebAuthn: o app guarda só a chave pública), com `userVerification` obrigatório, o que já vale como 2 fatores (posse do aparelho + biometria ou PIN). Google e GitHub entram **depois**, vinculados pelo e-mail verificado, nunca obrigatórios.
- O admin é criado **pelo dono**, no servidor (`api create-admin`), que emite um **convite de uso único** (256 bits, só o hash no banco, 15 minutos). Não existe rota HTTP para criar admin.
- Recuperação: alunos por e-mail (Brevo); **admin só por comando no servidor**.
- Entrega em 4 PRs: contas e convite (esta), passkeys, cadastro por e-mail e recuperação, provedores sociais. Migrações só aditivas (expand/contract); as colunas `idp_issuer/idp_subject` ficam até o `contract`.

## Consequências
- Menos infraestrutura e nenhuma credencial secreta de usuário para vazar; em troca, o app passa a implementar os fluxos de passkey, convite e recuperação (superfície de ataque própria, coberta por testes de ataque e limite de tentativas).
- O 2FA **não é obrigatório**: vem desligado e o administrador o liga na própria conta (`PUT /me/two-factor`; por ora só admin). Ligado, o login só abre sessão com 2 fatores comprovados: passkey com verificação vale; o Google não informa `amr`, então não vale.
- Perdeu a passkey: aluno recupera por e-mail; admin pelo servidor (exige acesso do dono).
- O Keycloak do `deploy/dev` fica só como dublê de provedor OIDC nos testes do adaptador (vale também para o Google).

## Alternativas descartadas
- **Senha no app:** o app saberia a senha de todos (recusado pelo dono).
- **Keycloak, Zitadel ou Logto próprios:** resolvem tudo, mas pedem mais RAM que a VM de tst e outro serviço para manter.
- **Só Google/GitHub:** amarra o cadastro a terceiros e não permite provar 2FA do admin.

## Fonte
`docs/features/2026-10-14-contas-proprias.md`; conversa de 14/10/2026.
