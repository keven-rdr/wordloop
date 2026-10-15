# Cadastro por e-mail e recuperação de senha (Brevo)

> Arquivo: `docs/features/2026-10-16-cadastro-email.md` · Issue: #__ · Branch: `feat/cadastro-email` · Estado: BUILD · PR 3 de 4 do plano de [contas próprias](2026-10-14-contas-proprias.md)

## Contexto
A tela "Criar conta" era só vitrine e o aluno não tinha como se cadastrar nem recuperar a senha. O ADR 0029 deixou isso para esta PR: cadastro aberto, confirmado por e-mail, e recuperação por e-mail só para alunos.

## Definições confirmadas
- **Cadastro aberto** a qualquer pessoa, confirmando o e-mail (resposta do dono em 15/10/2026).
- E-mail pelo **Brevo** (plano gratuito).
- O **admin nunca recebe e-mail** de cadastro nem de recuperação: ele se recupera pelo servidor (`deploy/scripts/reset-admin-access.sql`).
- Login com Google/GitHub criando ou vinculando conta fica para a PR 4.

## Decisões
| Tema | Decisão | Alternativa descartada / motivo |
|---|---|---|
| Cadastro | `POST /auth/register {email, locale}` cria uma conta **pendente** (`app_user` com e-mail e sem `email_verified_at`) e manda o link de confirmação (24 h). A resposta é **sempre 202** | criar a conta só depois do clique: o token de uso único precisa de um `user_id` |
| Sem enumeração | e-mail novo, pendente, já confirmado, de admin ou inexistente: mesma resposta (202, corpo vazio). Quem já tem conta recebe um e-mail "você já tem conta" com o caminho de login e um link de redefinição | responder "e-mail já cadastrado": revela quem tem conta |
| Recuperação | `POST /auth/password/forgot` manda um link (30 min) só a aluno com e-mail confirmado; admin, pendente e inexistente: 202 sem e-mail | recuperar admin por e-mail: o e-mail viraria a chave mestra (ADR 0028) |
| Definir a senha | a mesma `POST /auth/password/set` agora aceita o token de **convite, confirmação ou recuperação** (campo `token`, antes `invite`). Confirmação e recuperação marcam o e-mail como confirmado; trocar a senha encerra as outras sessões; token de recuperação de admin é recusado mesmo que exista | uma rota por tipo: três vezes o mesmo código |
| Segredo no link | `https://…/verify#<segredo>` e `/reset#<segredo>`: o fragmento não vai ao servidor nem no `Referer`, e a tela o apaga da barra de endereço | segredo na query string: vai para logs e histórico |
| Limites | no máximo 1 e-mail por minuto e 5 por hora por conta (excesso é descartado em silêncio), além do limite por IP de `/auth/*` | sem limite: dá para usar o app para spam |
| Envio | em segundo plano (a resposta não espera o Brevo e o tempo não revela se o e-mail existe); falha vai para o log, nunca para o cliente. Sem `BREVO_API_KEY`: 503 nas duas rotas e a tela avisa | falhar o cadastro por falha de e-mail: reabre o oráculo de existência |
| Idioma | e-mails em pt-BR e en-US (modelos embutidos em `identity/registration/templates`); idioma desconhecido cai em pt-BR | texto do e-mail em constantes no código |
| Anti-oráculo | o 202 não depende do e-mail existir, mas erro **interno** (banco fora do ar) responde 500: não depende da existência da conta | esconder falhas atrás de 202: o usuário acharia que o e-mail saiu |

## Limitações conhecidas
- Quem registra o e-mail de outra pessoa a cada ~1 minuto invalida o link que ela acabou de receber (o segredo só existe no e-mail; o banco guarda o hash). O estrago é só reenviar; o limite de 5 e-mails por hora também protege a caixa de entrada.
- O limite por e-mail lê e grava em dois passos: pedidos simultâneos podem passar do limite por minuto (não do limite por IP).
- "Esqueci a senha" não ajuda quem nunca confirmou o e-mail: nesse caso basta pedir o cadastro de novo (o link de confirmação é reenviado).
- O token de confirmação deixa de valer se o e-mail já foi confirmado por outro caminho.

## Fora desta PR
Limpeza periódica de contas pendentes e tokens vencidos (sem custo de segurança: pendente não entra); trocar a própria senha logado; Google e GitHub (PR 4).

## Checklist de progresso
- [x] `platform/mail`: Brevo (HTTP) e driver de log para desenvolvimento
- [x] `identity/registration`: cadastro, "já tem conta", recuperação, limites e modelos pt-BR/en-US
- [x] `identity/passwords`: token genérico (convite, confirmação, recuperação)
- [x] Contrato: `POST /auth/register`, `POST /auth/password/forgot`, campo `token`; handlers
- [x] Web: cadastro funcionando, "Esqueceu a senha?", páginas `/forgot`, `/verify` e `/reset`, textos pt-BR/en-US
- [x] Testes: Go (unidade, HTTP, Postgres real de ponta a ponta), Vitest e Playwright
- [x] Deploy: `MAIL_FROM`/`MAIL_FROM_NAME` no `api.env`, secret opcional `BREVO_API_KEY`
- [x] `npm run check` verde
- [ ] PR aberta (com aprovação do usuário)

## O que o dono precisa fazer para ligar em tst
1. Criar a conta no Brevo e verificar o domínio `wordloop.com.br` (DKIM e SPF no DNS).
2. Gerar uma chave de API (*SMTP & API → API Keys*) e salvar como secret `BREVO_API_KEY` no GitHub. **Não mande a chave no chat.**
3. Conferir `MAIL_FROM` em `deploy/tst/api.env` (precisa ser um remetente verificado no Brevo).

## Verificação
Automática: `go test ./...`, `go test -tags=integration ./...`, `npm run test`, `npm run test:e2e`.
Manual (dev): `MAIL_DRIVER=log APP_PUBLIC_URL=http://localhost:5173 api`, abrir `/signup`, pedir o link e copiar do log da API.
