# Fase 0 — Alinhamento

> Status: respostas recebidas (ver seção 6, que **prevalece** sobre as seções 3–5); faltam poucos dados. Data de referência: 2026-10-06.
> Rótulos: `[PREMISSA]` assumi; `[VERIFICAR]` não confirmei em fonte primária; `[DECISÃO SUA]` só o usuário decide.

## 1. O que verifiquei

**Contratos (somente leitura; nenhum `.env*` aberto, nenhum conteúdo copiado).** Conferi só estrutura e contagens do front:

| Item do Anexo A | Observado hoje | Obs. |
|---|---|---|
| `router.tsx` | 1.278 linhas | confere |
| `src/services` / `queries` / `schemas` | 53 / 50 / 58 arquivos | Anexo dizia 52 / 49 / 56: o repositório cresceu; a tese não muda |
| `CLAUDE.md` / `AGENTS.md` / `LLM.md` | 23,6 KB / 10,6 KB / 8 KB | confere: três arquivos sobrepostos |
| Imagem por ambiente | `Dockerfile`, `Sonar.Dockerfile`, `SonarPullRequest.Dockerfile`, `.env.hmg` na raiz | indício de build por ambiente `[VERIFICAR]` na Fase 4 |
| Pastas | `.azurepipelines`, `.crp-ai`, `.agents`, `e2e`, `scripts` | `scripts/` deve ter o `check-sonar-rules.mjs`; olho na Fase 4 |

**Versões e fatos externos** (consultas de hoje; só o registry do npm e o go.dev são fonte primária):

| Fato | Resultado | Status |
|---|---|---|
| Go | `go1.27.1` é o primeiro estável listado em https://go.dev/dl/?mode=json (1.27 previsto para ago/2026 segundo https://go.dev/doc/go1.27); 1.27 traz métodos genéricos, pacote `uuid` e novo engine JSON | confirmado |
| StyleX | `@stylexjs/stylex` **0.19.1** (https://registry.npmjs.org/@stylexjs/stylex/latest) — ainda 0.x | confirmado |
| Plugin Vite do StyleX | `@stylexjs/unplugin` **0.19.1**, export `./vite`, depende de `unplugin ^2.3.11` (https://registry.npmjs.org/@stylexjs/unplugin/latest) | confirmado; compatibilidade com Vite 8 `[VERIFICAR]` na Fase 3 |
| Vite / `@vitejs/plugin-react` | plugin-react **6.1.2**, usa `@rolldown/pluginutils`; Babel só via `@rolldown/plugin-babel` opcional + `babel-plugin-react-compiler` (https://registry.npmjs.org/@vitejs/plugin-react/latest). Confirma sua observação: React Compiler agora é opt-in e StyleX precisa do seu próprio transform | confirmado; teste prático na Fase 3 |
| MinIO | repositório arquivado em 25/04/2026 ([github.com/minio/minio](https://github.com/minio/minio)); imagens do Docker Hub removidas em 11/09/2026 segundo [youngju.dev](https://www.youngju.dev/blog/2026-07-17-minio-archived-garage-seaweedfs-ceph-rgw.en) (blog, não oficial) | arquivamento confirmado; remoção das imagens `[VERIFICAR]` |
| SonarQube Cloud Free | privados até 50k LOC, públicos sem limite; análise de branch/PR limitada à branch principal ([docs](https://docs.sonarsource.com/sonarcloud/administering-sonarcloud/managing-subscription/subscription-plans); [blog](https://www.sonarsource.com/blog/the-new-sonarqube-free-tier-is-here/)); plano para open source previsto para 2026 | confirma sua nota; plano OSS `[VERIFICAR]` |
| Astryx | Meta, beta público de 28/06/2026, MIT, 150+ componentes sobre StyleX ([BetterStack](https://betterstack.com/community/guides/ai/astryx-meta-design-system/), [OpenReplay](https://blog.openreplay.com/first-look-astryx-meta-design-system/)) | só fontes secundárias; repositório oficial `[VERIFICAR]` |

## 2. Correções da seção 10 — confirmo todas

| Ponto | Minha leitura |
|---|---|
| i18n | react-i18next com `useTranslation("ns")`; o problema a resolver é o registro central, não a API |
| Webhook | Web Push (VAPID) é o mecanismo; comparo com OneSignal/Ntfy na Fase 3 |
| Idioma da UI | pt-BR padrão + en-US desde o início; conteúdo = inglês; arquitetura aberta a novos locales |
| Lint | Biome + ESLint só para `@stylexjs/eslint-plugin` `[VERIFICAR]` |
| Imagem | uma imagem por versão, config em runtime |
| Storage | sem MinIO; volume + proxy (ou Garage/SeaweedFS se S3 for necessário) |
| StyleX | decidido; só riscos e mitigação |

## 3. Pontos que preciso levantar (franqueza)

1. **Hospedagem + RAM.** Três ambientes (dev local, tst, hmg) + Postgres ×2 + Keycloak + API + nginx + worker numa VM só. Keycloak sozinho costuma pedir ~0,7–1 GB `[VERIFICAR]`. Sem os dados da VM não dá para recomendar identidade (pergunta 6). Mitigação possível: tst e hmg compartilharem um Keycloak com realms separados, ou tst efêmero.
2. **tst e hmg na mesma VM de "produção".** Se não houver prd, hmg pode ser o ambiente que você realmente usa para estudar. Isso muda backup, dados reais e a política de aprovação.
3. **Repositório privado × público.** Público: Sonar sem limite, Actions grátis, e combina com conteúdo CC BY-SA. Privado: 50k LOC no Sonar, 2.000 min/mês, e o conteúdo derivado de CC BY-SA ainda exige compartilhar adaptações. **Recomendo público** `[DECISÃO SUA]`.
4. **Licença do código.** Conteúdo CC BY-SA 4.0 (NGSL) ≠ licença do código. Preciso saber se você aceita código MIT/Apache e conteúdo CC BY-SA separados no repositório.
5. **MVP.** Sua sugestão é grande para uma pessoa: login Google + SRS + painel + metas + deploy em dois ambientes + 300–500 imagens curadas. A curadoria de imagens costuma ser o gargalo real. Na Fase 6 vou propor MVP menor (ex.: ~150 palavras concretas com imagem, resto só com frase-contexto).
6. **Notas Again/Hard/Good/Easy × três botões.** Seu mapeamento (errou → Again) é plausível, mas FSRS usa 4 notas e a ilusão de competência (Koriat & Bjork) afeta a nota subjetiva; a Fase 1 trata disso.
7. **Multiusuário implica LGPD** (exportar/excluir, base legal, política de privacidade). Se for só você, parte do escopo some.
8. **Tempo disponível para construir.** Não perguntaste, mas dimensiona o roadmap: preciso de horas/semana.

## 4. Perguntas (padrão sugerido entre parênteses)

Responda "aceito os padrões" ou corrija só o que importa.

| # | Pergunta | Padrão |
|---|---|---|
| 1 | Meta semanal = dias de estudo (3/5/7) + desafios de sequência (7/15/30)? Lembretes = horários e frequência; retenção desejada fica em "avançado"? | sim |
| 2 | Haverá prd? | prd reservado para depois |
| 3 | Banco | PostgreSQL |
| 4 | Repositório público ou privado; monorepo? | público + monorepo |
| 5 | Sonar Cloud + trunk-based? | sim |
| 6 | Identidade: Keycloak ou auth própria em Go? | decidir na Fase 3, depois de ver a RAM da VM |
| 7 | Deploy: Compose + overlays + GitHub Environments, ou k3s + Argo CD? | Compose |
| 8 | Front: primitives + tokens StyleX (Astryx como referência) e roteamento por arquivo/manifesto? | sim |
| 9 | MVP inclui TTS do navegador, offline e nivelamento? | TTS e PWA offline básico: depois do núcleo; nivelamento: depois |
| 10 | Só você ou multiusuário desde o início? | multiusuário, com LGPD mínima |
| 11 | Imagens | bancos abertos + curadoria offline; IA só se couber no orçamento |
| 12 | Kit para agentes de IA enxuto, sem nada da empresa? | sim |
| 13 | Horas por semana que você consegue dedicar, e se há prazo desejado | `[PREENCHER]` |
| 14 | Licença do código | MIT, conteúdo CC BY-SA em diretório e licença separados |
| 15 | Seu ambiente local: Windows + Docker Desktop (ou WSL2)? | Windows + Docker Desktop |

**Dados de `[PREENCHER]` que ainda faltam:** nível atual de inglês; tempo diário disponível; dispositivos (Android/iPhone/PC); VM (SO, CPU, RAM, provedor, domínio).

## 5. Resumo de continuidade (para colar em chat novo)

Projeto: PWA mobile-first de vocabulário de inglês para falantes de pt-BR, com repetição espaçada guiada por feedback (Difícil/Médio/Fácil após cada resposta), imagens para formar conceito e progressão até leitura de páginas. Projeto pessoal, sem orçamento; também de estudo. Stack: API em Go (monólito modular, hexagonal por módulo, Postgres) + React 19 + TypeScript + Vite + StyleX (decidido; 0.19.1 hoje). CI/CD em GitHub Actions, Sonar, imagem única promovida entre tst/hmg com config em runtime, deploy na VM própria. Só planejamento, em 6 fases (0 alinhamento; 1 SRS; 2 pedagogia/conteúdo; 3 arquitetura/features; 4 engenharia; 5 UX; 6 roadmap/seed), parando ao fim de cada fase. Contratos (projeto de trabalho) é referência de convenções; nada dele é copiado.

Fatos confirmados em 2026-10-06: Go 1.27.1 estável; `@stylexjs/stylex` e `@stylexjs/unplugin` 0.19.1; `@vitejs/plugin-react` 6.1.2 sem Babel por padrão; MinIO arquivado em 25/04/2026; SonarQube Cloud Free limita análise de branch/PR à principal e privados a 50k LOC; Astryx beta (28/06/2026, MIT; fontes secundárias). Contratos conferido: `router.tsx` 1.278 linhas; services/queries/schemas com 53/50/58 arquivos; três arquivos de instrução sobrepostos.

Correções aceitas: i18n sem registro central; Web Push em vez de webhook; UI pt-BR + en-US; Biome + ESLint só para StyleX; imagem única; sem MinIO.

Pendente: respostas às perguntas 1–15 (seção 4) e dados `[PREENCHER]` (nível, tempo diário, dispositivos, VM). Recomendações iniciais: repo público monorepo, Sonar Cloud + trunk-based, Compose, multiusuário com LGPD mínima, identidade decidida na Fase 3 conforme RAM da VM. Próxima etapa: Fase 1 (SRS).

## 6. Decisões do usuário (prevalecem sobre o resto deste documento)

| Tema | Decisão |
|---|---|
| VM | Contabo Cloud VPS: 6 vCPU, 12 GB RAM, 100 GB NVMe (ou 200 GB SSD), 300 Mbit/s. SO e domínio ainda faltam. Cabe tst + prd + Keycloak com folga `[VERIFICAR]` na Fase 4 |
| Ambientes | **dev, tst, prd** (hmg foi removido). tst = promoção automática; prd = aprovação manual |
| Banco | PostgreSQL |
| Repositório | público, monorepo; já existe: `github.com/keven-rdr/wordloop` (1 commit em `main`; pasta local já vinculada) |
| Branches / Sonar | PRs para `develop` (nome convencional do Contratos; o usuário escreveu "developer": confirmar). Sonar Cloud Free analisa PR/branch **só da branch principal do projeto no Sonar**: configurar `develop` como branch principal lá, assim as PRs para `develop` são analisadas. PRs `develop → main` (release) ficam sem análise, aceitável porque o código já foi analisado `[VERIFICAR]` na Fase 4 |
| Identidade | aceitou "decidir na Fase 3"; com 12 GB, Keycloak passa a ser a recomendação provisória |
| Deploy | Docker Compose + overlays + GitHub Environments |
| Front | primitives + tokens StyleX; roteamento por arquivo ou manifesto |
| Escopo | **TTS, PWA offline e nivelamento entram no MVP** (risco de escopo registrado; a Fase 6 sequencia em incrementos) |
| Usuários | multiusuário (LGPD mínima) |
| Imagens | bancos abertos + curadoria offline |
| Kit de IA | sim, enxuto |
| Metas | sugestões padrão; ao fim de cada ciclo o app pergunta de novo, e a nova meta vale independentemente de a anterior ter sido cumprida |
| Licença | app open source; licença do código a definir (padrão proposto: MIT; conteúdo CC BY-SA separado) |
| Dev local | Windows + Docker Desktop |

### Lembretes = notificação-desafio (definição do usuário)

1. A notificação traz uma **pergunta curta** (ex.: `sort = ?`, `cat = ?`, `between = ?`, `I live in a hot country = ?`), a qualquer momento dentro das regras do usuário.
2. O toque abre o app **direto na pergunta** (deep link); o usuário responde.
3. Depois da resposta: opção **"explique a resposta"** e, se errou, **"entender por que errei"**; a **imagem entra aqui**, como contexto depois da tentativa de recuperação (coerente com Carpenter & Olson 2012 `[VERIFICAR]`).
4. A resposta atualiza o % de conhecimento; se for baixo, o item volta com **perguntas relacionadas** até atingir nível maior.

Implicações já identificadas:
- Web Push **não** permite digitar nem validar a resposta dentro da notificação de forma confiável (botões de ação não existem no iOS; não há campo de texto). A resposta acontece no app.
- "A qualquer momento" convive com anti-fadiga: janela de horários do usuário, máximo por dia, horário silencioso. Sorteio do instante dentro da janela.
- A pergunta vem de um item **vencido ou em aprendizado**, nunca de um item dominado fora do prazo.
- "Perguntas relacionadas" = o **mesmo item** em outro exercício/contexto (cloze, outra frase, imagem→palavra), **não** palavras semanticamente próximas, para evitar interferência (Tinkham 1993; Nakata & Suzuki 2019 `[VERIFICAR]`).
- O texto da pergunta aparece na tela de bloqueio: aceitável para vocabulário; opção de ocultar em Configurações.

### Ainda faltam

Nível atual de inglês; tempo diário disponível; dispositivos (Android/iPhone/PC); SO e domínio da VM; horas por semana para construir o projeto (e prazo, se houver); confirmar `develop` × `developer`; licença do código (MIT?).

### Resumo de continuidade (substitui o da seção 5)

Projeto `wordloop` (repositório antes chamado `poorAnki`): PWA mobile-first de vocabulário de inglês para falantes de pt-BR, com repetição espaçada guiada por feedback (Difícil/Médio/Fácil após cada resposta), imagens como contexto após a recuperação, e progressão até leitura de páginas. Open source, público, monorepo em `github.com/keven-rdr/wordloop`; multiusuário. Stack: API em Go (monólito modular, hexagonal por módulo, PostgreSQL) + React 19 + TypeScript + Vite + StyleX (0.19.1, decidido). Branches `develop`/`main`; Sonar Cloud Free com `develop` como branch principal; GitHub Actions; imagem única promovida entre ambientes dev/tst/prd com config em runtime; deploy em Contabo VPS (6 vCPU, 12 GB) com Docker Compose + overlays. Identidade: Keycloak provisório (decide na Fase 3). Entram no MVP: TTS, PWA offline e nivelamento (risco de escopo; sequenciar na Fase 6). Lembretes são notificações-desafio (pergunta curta → abre o app → resposta → explicação/imagem → % de conhecimento → itens relacionados), via Web Push (VAPID). Fatos de 2026-10-06: Go 1.27.1; StyleX/unplugin 0.19.1; plugin-react 6.1.2 sem Babel por padrão; MinIO arquivado (sem MinIO); Sonar Free limita análise à branch principal. Só planejamento, 6 fases com parada. Próxima: Fase 1 (SRS).

## 7. Atualização final da Fase 0 (respostas do usuário)

| Tema | Decisão / consequência |
|---|---|
| Nível do usuário | **Baixo** (vocabulário muito raso; não entende nem com legenda). `[PREMISSA]` ≈ A1 do CEFR. Consequências: o nivelamento precisa discriminar bem abaixo de ~1.000 palavras; conteúdo inicial = palavras de altíssima frequência com frases curtas e traduções PT-BR; explicações e UI em pt-BR; áudio (TTS) entra cedo, porque o gargalo declarado é compreensão auditiva |
| Tempo diário | Cerca de 60 min, **mas dinâmico**: o app pergunta ao usuário ao **iniciar** a meta; ao **fim da semana** reseta e pergunta de novo (a nova meta vale independentemente de ter cumprido a anterior) |
| Tempo semanal | É a **soma** do que foi estudado nos dias (60 min num dia, 20 em outro); o app só apresenta o total |
| Princípio de produto | O que importa não é o tempo, e sim **entrar, exercitar e melhorar**. A meta existe para motivar o retorno ao app, não para ser fixa. Métrica de sucesso sensível a melhora (ex.: palavras que passaram a ser lembradas, tendência de acerto, cobertura de vocabulário), e não "minutos cumpridos". Fica como requisito para o painel (Fase 3) e para as metas (Fase 5) |
| Dispositivos | Android, iPhone e PC. iPhone: Web Push só com o PWA instalado na Tela de Início (iOS ≥ 16.4) e permissão pedida por gesto; entra como teste obrigatório do Sprint 1 |
| VM | Contabo Cloud VPS (6 vCPU, 12 GB, NVMe) **ainda não comprada**. SO sugerido: Ubuntu LTS `[VERIFICAR]` versão. **Domínio**: necessário, pois Web Push e PWA exigem HTTPS válido; a Contabo não o fornece, então comprar um (ex.: `.com.br`) e apontar DNS. Porta 25 de saída costuma ser bloqueada em VPS `[VERIFICAR]`: usar provedor de e-mail transacional por API/SMTP 587 |
| Branch | `develop` (confirmado) |
| Licença | **MIT para o código** (decisão confirmada). MIT permite uso, cópia, modificação e uso comercial, exigindo só manter o aviso de copyright; para um app de aprendizado sem fins lucrativos não há obstáculo `[não é aconselhamento jurídico]`. O que exige cuidado é o **conteúdo**: NGSL e Tatoeba têm exigência de atribuição, NGSL é CC BY-SA 4.0 (compartilhar adaptações na mesma licença) e cada imagem tem sua própria licença. Por isso conteúdo e imagens ficam em diretório próprio com licença e atribuição por item (tabelas `sources`/`attributions`, Fase 2) |
| Horas para construir | A resposta foi sobre estudo do usuário, não sobre o desenvolvimento. `[PREMISSA]` ~8–10 h/semana, sem prazo; ajusto o roadmap da Fase 6 se você corrigir |

## 8. Fase 0 encerrada

Nada grande pendente. Pendências pequenas (não bloqueiam a Fase 1): SO/domínio da VM (após a compra); horas semanais para construir (assumi 8–10 h).
