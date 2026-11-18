# Pronto para produção (Sprint 6)

> Arquivo: `docs/features/2026-11-18-producao-beta.md` · Issue: #__ · Branch: `feat/lgpd-api` · Estado: BUILD (PR 1 pronta; PRs 2 a 4 pendentes)

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
| Como a exportação é montada | uma consulta no Postgres devolve o JSON (`jsonb_build_object`), com as colunas **sempre listadas**; a API só acrescenta `schema_version` e `generated_at` | mapear linha a linha em Go: muito código para conferir e fácil esquecer uma coluna; `SELECT *`: um segredo entra no dia em que uma coluna nova nascer |
| Encerrar a sessão depois de apagar | a exclusão apaga as linhas de sessão (cascata); o app chama `POST /auth/logout` em seguida para limpar o cookie | o módulo `privacy` conhecer os cookies do `identity` |
| Formato da exportação | JSON, entregue na hora | e-mail com link: depende de SMTP, que ainda não existe |

## Checklist de progresso
- [x] Roadmap: Sprint 5 concluído e Sprint 6 em andamento
- [x] PR 1: LGPD na API (exportar e excluir: migração 00018, módulo `privacy`, contrato)
- [ ] PR 2: LGPD na web e páginas de privacidade e termos
- [ ] PR 3: backup, restauração, prd endurecido e deploy de prd (como código e doc)
- [ ] PR 4: Google com vínculo e e-mails por SMTP (prontos para ligar)
- [ ] `npm run check` verde em cada PR
- [ ] PRs abertas (com aprovação do usuário)

## Verificação
- **PR 1:** `npm run check` verde (golangci-lint, estrutura, arch-lint com os componentes de `privacy`, `go test -short`, web). Testes unitários cobrem as regras de exclusão (e-mail sem diferenciar maiúsculas nem espaços, parte do e-mail, conta sem e-mail, administrador, login de 15 min contra 15 min e 1 s, e qual aviso vem primeiro), os casos de uso (nada é apagado quando uma regra falha; falhas sobem; conta inexistente) e o contrato HTTP (204, 400 com e sem o cabeçalho de proteção, 401, 403 e 428; o arquivo exportado traz `Content-Disposition`, versão e data; apagar duas vezes não vaza nada).
  - **Não rodei** os testes de integração com Postgres (sem Docker aqui): compilam com `-tags integration`; o CI roda. Cobrem a migração 00018 e o que mais importa: a exportação **nunca** traz um segredo (nem por nome de coluna nem por valor, inclusive em hexadecimal), não traz dado de outra pessoa e vem completa mesmo para quem nunca estudou; a exclusão **não deixa linha em nenhuma tabela com `user_id`** (a lista vem do próprio banco, então vale para tabelas futuras), não mexe nos dados de outra pessoa, registra a exclusão sem dado pessoal (o teste confere as colunas da tabela), é "tudo ou nada" e o `review_log` **continua imutável** fora da exclusão (inclusive declarando outra pessoa ou tentando `UPDATE`).
  - A tela de exportar e excluir é a PR 2: por ora os dois endpoints só existem na API.
  - **No CI da PR 1:** o gitleaks tratou um valor falso do teste (um "canário") como segredo: renomeado, com o falso positivo do commit anterior no `.gitleaksignore`; o Sonar reprovou por cobertura de código novo (72,3% contra 80%), e a ligação dos módulos em `cmd/api/learner.go` ganhou testes (os adaptadores de sessão e a montagem de cada módulo), assim como os caminhos de erro do handler. O e2e das preferências tinha uma corrida (lia o `localStorage` antes de a resposta do `PUT` aplicar o tema) e passou a esperar o valor.
