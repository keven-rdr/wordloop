# ADR 0029 — Senha guardada só como hash Argon2id, ao lado da passkey

**Status:** aceita (2026-10-16). **Revisa o [ADR 0028](0028-contas-proprias-sem-senha-passkeys-depois-google-github.md)** apenas no ponto "nenhuma senha, em lugar nenhum": o resto (contas próprias, convite do admin, Google e GitHub só vinculados) continua valendo.

## Contexto
O ADR 0028 escolheu entrar só por passkey. Na prática o dono não quer depender da passkey para entrar como admin e pediu **e-mail e senha**, mantendo a regra de fundo: "o app não pode saber a senha de ninguém". Isso é possível sem guardar a senha: o banco guarda só uma marca irreversível dela (hash com sal), e a senha em si só existe na memória durante o pedido.

## Decisão
- Entrar por **e-mail e senha** (`POST /auth/password/login`) e/ou por **passkey**; as duas formas coexistem. A senha é opcional por conta.
- A senha vira hash **Argon2id** (19 MiB, 2 passadas, 1 thread, sal de 16 bytes) em `app_user.password_hash`. No máximo 2 hashes ao mesmo tempo, porque a VM de tst tem 1 GB.
- Política (NIST): 12 a 128 caracteres, diferente do e-mail, sem um caractere repetido; sem regras de símbolos.
- A senha é definida pelo **convite** de uso único que o admin já recebe (`POST /auth/password/set`), e o convite é consumido na mesma transação. O esquecimento de senha do admin se resolve por SQL/comando no servidor (`deploy/scripts/reset-admin-access.sql`), nunca só por e-mail.
- Defesas: toda recusa é um 401 idêntico (sem revelar quais e-mails existem), com um hash "isca" conferido quando a conta não existe (tempo parecido); a tentativa é **contada antes** de conferir a senha, de forma atômica no banco (uma rajada de requisições simultâneas não passa de 5 conferências); a 5ª ainda confere e, se errar, bloqueia as seguintes por 30 s, dobrando até 15 min (tentativas com mais de 15 min de idade não contam); o bloqueio também responde 401 e o acerto zera tudo; limite por IP já existente em `/auth/*`; cabeçalho anti-CSRF; senha nunca em log, URL ou resposta.
- Senha sozinha **não prova 2 fatores**. Conta que ligou o 2FA ([2FA opcional](../features/2026-10-16-2fa-opcional.md)) recebe 403 no login por senha e entra pela passkey.

## Consequências
- O app volta a ter um segredo de usuário para proteger (o hash). Ele é útil a um atacante só com muito custo (Argon2id), e o bloqueio por tentativas limita o ataque online.
- O bloqueio por conta permite que alguém trave de propósito a conta do admin por alguns minutos (negação de serviço leve). A saída é a passkey e, em último caso, o reset por SQL.
- Aluno recupera a senha por e-mail e se cadastra por e-mail, com confirmação ([cadastro por e-mail](../features/2026-10-16-cadastro-email.md)); admin só se recupera pelo servidor.

## Alternativas descartadas
- **Manter só passkey:** é o que o dono recusou nesta data.
- **Senha em provedor externo (Keycloak, Auth0...):** amarra o app a um terceiro, já recusado no ADR 0028.
- **bcrypt/scrypt/PBKDF2:** Argon2id é a recomendação atual da OWASP para senhas novas.

## Fonte
`docs/features/2026-10-16-senha-do-admin.md`; conversa de 16/10/2026.
