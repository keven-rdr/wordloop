# ADRs — registro de decisões estruturais

Gerados por `gen_adrs.py` em 2026-10-06. Cada ADR é curto; o raciocínio completo está no documento da fase indicado.

| # | Decisão | Fonte |
|---|---|---|
| 0001 | [Monólito modular em Go](0001-monolito-modular-em-go.md) | `fase-3/03-arquitetura-features.md` |
| 0002 | [Pacote por funcionalidade e limites medidos no CI](0002-pacote-por-funcionalidade-e-limites-medidos-no-ci.md) | `fase-3/03b-estrutura-pastas.md` |
| 0003 | [`srs` como biblioteca pura com Strategy (`simple_v1`, `fsrs6`)](0003-srs-como-biblioteca-pura-com-strategy-simple-v1.md) | `fase-1/01-srs.md` |
| 0004 | [Domínio (derivado de S) como '% de conhecimento' na UI; R só para fila e notificação](0004-dominio-derivado-de-s-como-de-conhecimento-na-ui.md) | `fase-1/01-srs.md` |
| 0005 | [`review_log` imutável e estado derivado por replay](0005-review-log-imutavel-e-estado-derivado-por-replay.md) | `fase-1/01-srs.md; fase-3/schema-learning.sql` |
| 0006 | [OpenAPI 3.0 primeiro; servidor `oapi-codegen`, cliente `hey-api`](0006-openapi-3-0-primeiro-servidor-oapi-codegen.md) | `fase-3/03-arquitetura-features.md; fase-3/openapi-sketch.yaml` |
| 0007 | [pgx + sqlc + goose](0007-pgx-sqlc-goose.md) | `fase-3/03-arquitetura-features.md` |
| 0008 | [Fila como outbox no PostgreSQL](0008-fila-como-outbox-no-postgresql.md) | `fase-3/03-arquitetura-features.md` |
| 0009 | [Eventos em processo; `ReviewCompleted` na mesma transação](0009-eventos-em-processo-reviewcompleted-na-mesma.md) | `fase-3/03-arquitetura-features.md` |
| 0010 | [Um Keycloak com dois realms (tst e prd)](0010-um-keycloak-com-dois-realms-tst-e-prd.md) | `fase-3/03-arquitetura-features.md (nota de decisões)` |
| 0011 | [API como BFF OIDC com cookie `httpOnly`](0011-api-como-bff-oidc-com-cookie-httponly.md) | `fase-3/03-arquitetura-features.md; fase-3/03c-diagramas.md` |
| 0012 | [Lembretes por Web Push nativo (VAPID)](0012-lembretes-por-web-push-nativo-vapid.md) | `fase-3/03-arquitetura-features.md` |
| 0013 | [StyleX com Base UI e tokens próprios](0013-stylex-com-base-ui-e-tokens-proprios.md) | `fase-3/03d-design-system-stylex.md` |
| 0014 | [TanStack Router por arquivo e `feature.ts` por `import.meta.glob`](0014-tanstack-router-por-arquivo-e-feature-ts-por.md) | `fase-3/03-arquitetura-features.md` |
| 0015 | [i18n por namespace carregado sob demanda, com tipos gerados](0015-i18n-por-namespace-carregado-sob-demanda-com.md) | `fase-3/03-arquitetura-features.md` |
| 0016 | [PWA offline com fila idempotente e pacote de estudo local](0016-pwa-offline-com-fila-idempotente-e-pacote-de.md) | `fase-3/03-arquitetura-features.md` |
| 0017 | [Monorepo público; `develop` e `main`; merge por squash](0017-monorepo-publico-develop-e-main-merge-por-squash.md) | `fase-4/04-engenharia.md` |
| 0018 | [release-please por componente; build uma vez e promoção por re-etiqueta](0018-release-please-por-componente-build-uma-vez-e.md) | `fase-4/04-engenharia.md` |
| 0019 | [Deploy por SSH com comando forçado e GitHub Environments](0019-deploy-por-ssh-com-comando-forcado-e-github.md) | `fase-4/04-engenharia.md; fase-4/deploy/deploy-skeleton.md` |
| 0020 | [Docker Compose em uma VM Contabo: três projetos, um Postgres, Caddy](0020-docker-compose-em-uma-vm-contabo-tres-projetos-um.md) | `fase-4/deploy/deploy-skeleton.md` |
| 0021 | [SonarQube Cloud Free com `develop` como branch principal](0021-sonarqube-cloud-free-com-develop-como-branch.md) | `fase-4/04-engenharia.md` |
| 0022 | [Imagens: OpenMoji como base, fora do Git, com manifesto de licença](0022-imagens-openmoji-como-base-fora-do-git-com.md) | `fase-2/02-pedagogia-conteudo.md` |
| 0023 | [Cobertura de vocabulário por texto como métrica de leitura](0023-cobertura-de-vocabulario-por-texto-como-metrica.md) | `fase-2/02-pedagogia-conteudo.md` |
| 0024 | [Licenças: código MIT e conteúdo CC BY-SA 4.0](0024-licencas-codigo-mit-e-conteudo-cc-by-sa-4-0.md) | `fase-2/02-pedagogia-conteudo.md; fase-4/04-engenharia.md` |
| 0025 | [Kit de instruções para IA: `AGENTS.md` canônico, hook e skill](0025-kit-de-instrucoes-para-ia-agents-md-canonico-hook.md) | `fase-4/04-engenharia.md; fase-4/kit-ia/` |
| 0026 | [Experiência: sem XP, ranking nem mascote; sequência por revisão; barra de 4 itens](0026-experiencia-sem-xp-ranking-nem-mascote-sequencia.md) | `fase-5/05-ux-ui-fluxos.md` |
