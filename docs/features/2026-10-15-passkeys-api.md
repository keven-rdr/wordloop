# Passkeys na API: registrar com o convite, entrar e gerenciar

> Arquivo: `docs/features/2026-10-15-passkeys-api.md` · Issue: #__ · Branch: `feat/passkeys-api` (PR 2a de 4; as telas do app são a PR 2b) · Estado: VERIFY

## Contexto
Segue o [ADR 0028](../adr/0028-contas-proprias-sem-senha-passkeys-depois-google-github.md) e a [PR 1 (contas próprias)](2026-10-14-contas-proprias.md): o app **não guarda senha**. Entrar é por **passkey** (WebAuthn): o servidor guarda só a **chave pública**; a chave privada nunca sai do aparelho e é liberada com biometria ou PIN (`userVerification: required`), o que também é o 2FA do admin.

Esta PR entrega só a **API**. O convite emitido por `api create-admin` passa a servir: abre o registro da primeira passkey e já entra logado. As telas (aceitar convite, "Entrar com passkey", "minhas passkeys") vêm na PR 2b.

## Definições confirmadas
- Passkeys + cadastro no app, depois Google e GitHub (decisão do dono, 14/10/2026).
- Admin só entra com 2 fatores comprovados: a passkey com verificação do usuário vale como 2 fatores.
- O app nunca recebe, guarda nem registra senha.

## Decisões
| Tema | Decisão | Motivo |
|---|---|---|
| Biblioteca | `github.com/go-webauthn/webauthn` v0.18.2, isolada no adaptador `identity/adapters/webauthn`; os casos de uso falam com uma interface (`Ceremony`) | trocar ou testar sem a biblioteca |
| Credenciais descobríveis | login **sem digitar usuário** (o aparelho mostra as passkeys do site) | menos atrito, nada a enumerar |
| Desafio | **uso único e no servidor**: tabela `auth_challenge` (hash do id, estado, 5 min) e um cookie `httpOnly` só com o id; `DELETE … RETURNING` ao usar | um cookie selado e sem estado poderia ser repetido |
| Convite | validado ao começar e **consumido junto com a gravação da passkey, na mesma transação**; usado, expirado ou inventado dá 401 igual | o convite não queima se a gravação falha; sem pista do motivo |
| Sessão após convite | registrar com o convite **já abre a sessão** (passkey com UV = MFA verdadeiro) | quem acabou de criar a passkey provou o aparelho |
| Segurança da passkey | atestado `none`; credencial repetida é recusada; passkey já cadastrada fica na lista de exclusão do aparelho; contador regressivo (clone) recusa o login e é registrado em log; no máximo 10 passkeys por conta | padrão da especificação |
| Apagar passkey | não deixa a conta sem nenhuma forma de entrar; **admin nunca fica sem passkey** (409). A checagem e o `DELETE` rodam **numa transação com a linha do usuário travada** (`SELECT … FOR UPDATE`), então dois apagamentos simultâneos não furam a regra | recuperação do admin é só pelo servidor |
| Limite de tentativas | `platform/ratelimit` por IP em `/api/v1/auth/*` (30/min, rajada de 15), 429 com `Retry-After`; corpo de pedido limitado a 64 KB. O IP vem do `X-Forwarded-For` **só com `API_TRUST_PROXY=true`** (ligado em tst, atrás do Caddy); sem isso o cabeçalho é ignorado e ninguém burla o limite trocando-o | a API não tinha nenhum limite |
| Configuração | `WEBAUTHN_RP_ID` e `WEBAUTHN_ORIGIN`; sem eles as rotas de passkey respondem 503 e a API sobe normal | mesmo padrão do OIDC |
| `create-admin` | passa a considerar passkeys ao decidir se o admin "já tem como entrar" | não reemitir convite de quem já registrou |

## Verificado com Postgres de verdade
O Docker local estava desligado, então os testes de integração rodaram contra o **PGlite** (motor PostgreSQL) servido por TCP: o fluxo completo (`migrate` → `create-admin` → convite → registrar passkey → `GET /me` admin → convite reutilizado 401 → login por passkey → listar → apagar a última passkey do admin 409) e 12 testes do adaptador passaram. Isso achou um erro real (apagar passkey de usuário inexistente voltava erro de banco em vez de 404). **Não rodaram aqui** (precisam de várias conexões, o CI as roda): a corrida de dois apagamentos simultâneos e a conferência de `create-admin` depois da passkey.

## Fora desta PR
Telas e roteador do app (PR 2b); cadastro por e-mail e recuperação (PR 3); Google e GitHub (PR 4); renomear passkey.

## Checklist de progresso
- [x] Contrato (`openapi.yaml` + `npm run gen`): 6 operações e 429
- [x] Migração 00005 (aditiva) e consultas `sqlc`
- [x] `app`: `SessionOpener` reutilizável com a regra do 2FA do admin
- [x] `internal/identity/passkeys`: casos de uso e testes
- [x] Adaptador `webauthn` testado com um autenticador de software (assinaturas reais)
- [x] Adaptador `postgres` (credenciais, desafios, convite em transação)
- [x] HTTP: handlers, cookie do desafio, limite de tentativas, ligação em `cmd/api`
- [x] Testes de ataque "estilo Postman"
- [x] `deploy/tst/api.env`, `docs/ci-cd.md`
- [x] `npm run check` verde
- [ ] PR aberta (com aprovação do usuário)

## Verificação
Automática: unidade (casos de uso com ceremony falsa), cerimônia real com autenticador de software (registrar, entrar, UV ausente, origem errada, contador regressivo, `userHandle` trocado), HTTP forjado (convite reutilizado/expirado, desafio repetido, credencial de outro usuário, CSRF, 429), SQL num Postgres real (PGlite). Integração (Postgres real) no CI. Manual: depois da PR 2b, `create-admin` → abrir o link → criar a passkey → entrar.
