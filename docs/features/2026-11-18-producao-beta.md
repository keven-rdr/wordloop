# Pronto para produção (Sprint 6)

> Arquivo: `docs/features/2026-11-18-producao-beta.md` · Issue: #__ · Branch: `feat/lgpd-api` · Estado: SPEC (requisitos confirmados; PR 1 em construção)

## Contexto
O Sprint 5 fechou o ciclo de estudo (progresso, metas, preferências). O Sprint 6 do [roadmap](../fase-6/06-roadmap-backlog.md) leva o app a **produção beta**: login Google com vínculo por e-mail (I-02, I-03), e-mails por SMTP (I-04), **LGPD: exportar e excluir** (I-05), privacidade e termos (I-06), backup com restauração (O-01), prd endurecido (O-02) e o release `0.x` em prd (F-06). Critério: "você estuda em prd; backup restaurado em banco descartável; exclusão apaga tudo".

**O roadmap está defasado em um ponto:** ele fala em "remover o usuário no Keycloak", mas o [ADR 0028](../adr/0028-contas-proprias-sem-senha-passkeys-depois-google-github.md) **substituiu o Keycloak por contas próprias** (passkey, senha com Argon2id, Google e GitHub vinculados por `auth_identity`). Apagar a conta é, portanto, apagar linhas do nosso banco; não há servidor de identidade externo a limpar. O Keycloak do `deploy/dev` é só um dublê de provedor OIDC nos testes.

## Perguntas e respostas dos requisitos
| # | Pergunta | Recomendação | Resposta |
|---|---|---|---|
| 1 | Em que pé está a infra (VM, domínio, cliente OAuth do Google, Brevo, R2)? | Nada ainda: entregar o que dá para fazer e testar sem essas contas; o resto sai pronto para ligar, com passo a passo | **Nada ainda** (18/11/2026) |
| 2 | Em que ordem? | Dados e privacidade primeiro (LGPD e páginas), depois backup/prd, depois Google e SMTP | Recomendação aceita (18/11/2026) |
| 3 | Como funciona a exclusão? | Imediata, com confirmação (digitar o e-mail + login recente); sem volta | Recomendação aceita (18/11/2026) |

## Definições confirmadas
- **PR 1 — LGPD na API (I-05):** módulo novo `internal/privacy` (domain · app · adapters), porque exportar e excluir cruzam várias tabelas e não pertencem a um só módulo.
  - `GET /me/export`: JSON completo da pessoa, como download (`Content-Disposition`). Traz perfil (nome, e-mail, datas), vínculos de login (provedor e data), passkeys (nome e datas), sessões (aparelho e datas), preferências, metas, dias de estudo, sessões de estudo, cartões e o **log de revisões inteiro**. **Nunca** traz segredo: hash de senha, chave pública da passkey, tokens e hashes de sessão ficam de fora.
  - `POST /me/erase` com `{ confirmation }`: apaga tudo na hora e encerra a sessão. Exige (a) a confirmação igual ao **e-mail da conta** (sem diferenciar maiúsculas), (b) **login recente** (autenticada há no máximo 15 minutos), (c) conta que **não seja de administrador** (admin só se apaga por comando no servidor, como já é para criar e recuperar). Respostas: 204; 400 confirmação errada; 401 sem sessão; 403 admin; 428 login antigo.
  - O `review_log` é imutável por gatilho. A exclusão usa uma regra **estreita**: o gatilho passa a permitir `DELETE` só quando a transação declara `wordloop.erase_user = <id da linha>` (`SET LOCAL`); `UPDATE` continua proibido e qualquer outro `DELETE` também. Sem travar a tabela.
  - Fica um registro **sem dado pessoal** de que houve exclusão (`account_erasure`: quando, e quantas respostas havia), para prestação de contas.
- **PR 2 — LGPD na web (I-05, I-06):** tela "Privacidade" nas Configurações (exportar com um toque; excluir com confirmação e aviso claro de que não tem volta), e páginas públicas de **privacidade** e **termos**, linkadas no login, no cadastro e nas Configurações. Os textos são **rascunhos técnicos para revisão jurídica** (eu não sou advogado): descrevem o que o app de fato faz com os dados.
- **PR 3 — O-01, O-02, F-06:** backup `pg_dump` criptografado (`age`) para o R2 e script de restauração testado em banco descartável; endurecimento da VM de prd (`ufw`, SSH por chave, `fail2ban`, limites de memória) e deploy de prd (compose, Caddy, versões por digest, aprovação). **Tudo como código e documentação**: nada é provisionado sem as contas do usuário.
- **PR 4 — I-02, I-03, I-04:** Google com vínculo por e-mail confirmado e e-mails por SMTP (Brevo), prontos para ligar por configuração. Os testes automáticos rodam contra dublês; a verificação real exige as credenciais do usuário.

## Decisões
| Tema | Decisão | Alternativa descartada / motivo |
|---|---|---|
| Quem é dono de exportar e excluir | módulo próprio `privacy`, que lê e apaga nas tabelas dos outros módulos por consultas suas (direito do titular atravessa o sistema inteiro) | pôr em `identity`: ela teria de conhecer cartões, metas e log |
| Apagar o log imutável | gatilho com exceção estreita (`SET LOCAL wordloop.erase_user`) | `ALTER TABLE … DISABLE TRIGGER`: trava a tabela para todos e abre a porta para qualquer `DELETE` |
| Confirmação | digitar o e-mail da conta + login de até 15 min | só um botão: erro de toque apaga tudo; pedir senha: a conta pode não ter senha (passkey/Google) |
| Admin | não se apaga pela API | apagar o único admin trancaria o sistema |
| Registro da exclusão | `account_erasure` sem id nem e-mail da pessoa | guardar o id: deixaria a pessoa identificável depois de pedir esquecimento |
| Formato da exportação | JSON, entregue na hora | e-mail com link: depende de SMTP, que ainda não existe |

## Checklist de progresso
- [x] Roadmap: Sprint 5 concluído e Sprint 6 em andamento
- [ ] PR 1: LGPD na API (exportar e excluir)
- [ ] PR 2: LGPD na web e páginas de privacidade e termos
- [ ] PR 3: backup, restauração, prd endurecido e deploy de prd (como código e doc)
- [ ] PR 4: Google com vínculo e e-mails por SMTP (prontos para ligar)
- [ ] `npm run check` verde em cada PR
- [ ] PRs abertas (com aprovação do usuário)

## Verificação
(a preencher a cada PR)
