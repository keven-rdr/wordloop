# Sessão segura e proteção do admin

> Arquivo: `docs/features/2026-10-14-sessao-segura-admin.md` · Issue: #__ · Branch: `feat/sessao-segura` (PR 1) e `feat/admin-auditoria` (PR 2) · Estado: VERIFY

## Contexto
O login (BFF) já existe: cookie `wl_session` de 30 dias fixos, sessão no banco. Isso é aceitável para aluno, mas **não para um admin**: quem roubar o cookie dele age como admin por 30 dias. O dono do projeto quer (1) impedir a tomada da conta de admin (2FA) e (2) impedir que alguém, com Postman ou outra ferramenta, **forje** sessão, papel ou identidade. O ataque considerado é o de pedido forjado (cookie inventado, `role` no corpo, id de outro usuário, CSRF), mais o roubo do cookie.

## Definições confirmadas (pelo dono, em 14/10/2026)
Lista aprovada ("quero isso"): sessão de admin curta e a de aluno com prazo por inatividade; 2FA do admin e confirmação recente nas ações perigosas; papel guardado no banco e conferido a cada ação; listar e derrubar sessões; usar o refresh token para confirmar a conta; auditoria das ações de admin; prefixo `__Host-` no cookie.

## Divisão em duas PRs
| PR | Conteúdo |
|---|---|
| **1 (esta)** | papel no banco; prazos absoluto e por inatividade por papel; admin só entra com 2FA; `__Host-` no cookie; `GET /me` com `role`; listar e derrubar sessões (uma ou todas) |
| **2** | primeira rota de admin (consulta da auditoria); guarda `RequireAdmin` com "login recente" (step-up); tabela de auditoria; revalidar a conta pelo refresh token; encerrar sessões quando o provedor avisar troca de senha (back-channel logout) |

## Decisões da PR 1
| Tema | Decisão | Alternativa descartada / motivo |
|---|---|---|
| Prazos (valores iniciais, ajustáveis em `domain/session_policy.go`) | aluno: **30 dias** absolutos e **7 dias** sem uso; admin: **8 horas** absolutas e **30 minutos** sem uso | sessão única de 30 dias: roubo dura demais |
| Onde vale o prazo | calculado **a cada pedido** com o papel **atual** do banco: quem vira admin passa a ter o prazo curto na hora | gravar o papel na sessão: promoção não teria efeito até o próximo login |
| Papel | coluna `app_user.role` (`user` ou `admin`), padrão `user`. **Só muda por SQL** feito pelo dono; nenhuma rota, corpo de pedido ou claim do provedor altera o papel | papel vindo do token: o provedor externo não é fonte de autorização |
| 2FA do admin | no login, se o usuário é admin e o ID token não prova 2 fatores (`amr` com `mfa`, `otp` ou `hwk`), **o login é recusado (403)** e nenhuma sessão nasce | confiar só na configuração do provedor: falha silenciosa se alguém desligar o 2FA lá |
| Cookie | `__Host-wl_session` (Secure, Path `/`, sem Domain) e `__Secure-wl_oidc`; sem o prefixo só em http (dev local) | prefixo sempre: Safari recusa cookie Secure em `http://localhost` |
| Sessões | cada sessão tem um `public_id` (uuid) próprio, **diferente** do segredo; `GET /me/sessions`, `DELETE /me/sessions/{id}`, `DELETE /me/sessions` (todas) | expor o hash do id: desnecessário |
| Dono da sessão | revogar sessão de outro usuário responde **404** (igual a "não existe"), nunca 403 | 403 revela que o id existe |
| CSRF | os dois `DELETE` exigem `X-Requested-With: XMLHttpRequest` (como o logout) além do `SameSite=Lax` | só o SameSite |
| Escrita por pedido | `last_used_at` só é gravado se passou mais de 1 min do último (poupa o Neon) | gravar em todo pedido |

## Por que o Postman não escala privilégio aqui
- A identidade vem **só** do cookie (id aleatório de 256 bits; o banco guarda o hash). Cookie inventado ou adulterado: 401.
- O contrato não tem campo de papel nem de id de usuário em nenhum corpo. O servidor gerado ignora campos não declarados, e os testes mandam `role: admin` e `user_id` de outro para provar.
- Ids de sessão e de usuário em caminho/consulta são sempre filtrados pelo dono da sessão (anti-IDOR).
- Pedido de outro site não leva o cookie (`SameSite=Lax`) e, mesmo que levasse, os `DELETE` exigem um cabeçalho que só o nosso JavaScript envia.

## Fora desta PR (explícito)
- Limite de tentativas (rate limit) nas rotas de login: PR própria (Caddy e/ou middleware).
- "Sair de todos os aparelhos" automático ao trocar a senha: depende de back-channel logout do provedor (PR 2).
- **Limpeza de sessões vencidas e limite de sessões por usuário** (hoje a linha só some quando a sessão é usada ou derrubada): PR 2, junto com um expurgo periódico.
- **Uma consulta só (JOIN) para sessão + usuário** em vez de duas leituras por pedido: otimização para quando houver medição; não muda o comportamento.
- **Corpo e código no 403 do admin sem 2FA** (catálogo de Problem Details): quando o catálogo existir.
- `Principal.AuthenticatedAt` e `MFA` já são gravados (a guarda de "login recente" da PR 2 os lê).
- **Regra de 2 fatores:** `amr` com `mfa`, ou `pwd` mais `otp`/`hwk`. Um método sozinho (`hwk` ou `otp`) **não** basta. O `amr` e o `auth_time` são lidos com tolerância: valor estranho nunca derruba o login, só deixa de provar 2 fatores.
- Comprovar `amr` no provedor escolhido: `[VERIFICAR]`. No Keycloak e nos provedores gratuitos a presença do `amr` depende de configuração; **se o provedor não enviar `amr`, o admin não consegue entrar** (falha fechada, de propósito).

## Como virar admin (manual, por enquanto)
Depois do primeiro login da conta, com 2FA ligado no provedor, no banco:
`UPDATE app_user SET role = 'admin' WHERE id = '<uuid do /me>';`
(Rever o passo na PR 2, quando houver auditoria.)

## Checklist de progresso
- [x] Contrato: `openapi.yaml` (+ `npm run gen`): `role` em `Me`, 403 no callback, 3 operações de sessões
- [x] Migração 00003 (aditiva): `app_user.role`, `auth_session.public_id/authenticated_at/mfa`
- [x] `domain`: `Role`, `SessionPolicy`, `Principal`; `Identity` com `AuthTime` e `MFA`
- [x] `app`: `Authenticate` (prazos por papel), `ListSessions`, `RevokeSession`, `RevokeAllSessions`, recusa de admin sem 2FA
- [x] Provedor OIDC lê `amr` e `auth_time`
- [x] HTTP: cookies com prefixo, handlers das sessões
- [x] Testes: unidade e HTTP "como no Postman" (rodados). Integração (Postgres e Keycloak reais) **compila mas só roda no CI**: o Docker local estava desligado; o SQL novo foi conferido num Postgres de verdade (PGlite)
- [x] `npm run check` verde
- [ ] PR aberta (com aprovação do usuário)

## Verificação
Automática: `go test ./...` (prazos, 2FA, anti-IDOR, cookie forjado) e `go test -tags=integration ./...` (Postgres e Keycloak reais, no CI). Manual em tst: login, `GET /me` com `role`, `GET /me/sessions`, derrubar uma sessão e conferir 401 no outro navegador.
