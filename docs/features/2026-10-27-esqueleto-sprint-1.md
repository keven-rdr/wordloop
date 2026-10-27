# Esqueleto ponta a ponta (Sprint 1): fechamento

> Arquivo: `docs/features/2026-10-27-esqueleto-sprint-1.md` · Branch: `docs/sprint-1-concluido` · Estado: DONE (com 2 exceções registradas)

## Contexto
O Sprint 1 do [roadmap](../fase-6/06-roadmap-backlog.md) pedia o *walking skeleton*: repositório com o kit de IA → `api-ci` e `web-ci` com gate → imagem única → deploy automático em tst → login → tela "Olá" com a versão visível → um E2E de fumaça. Critério de aceite: **PR → CI verde → tst atualizado → versão visível; rollback por redeploy testado**.

Este documento fecha o sprint, item por item, com a evidência. Depende das PRs de fechamento do Sprint 1 (smoke test, rollback, Lighthouse e Trivy) já mescladas.

## Itens do backlog
| Item | Entrega | Evidência |
|---|---|---|
| F-01 Repositório com kit de IA | `npm run check`, hook `guard.mjs`, lefthook, `develop` padrão | `AGENTS.md`, `scripts/check.mjs`, `.claude/hooks/guard.mjs` |
| F-02 `api-ci` com gate | lint, testes com `-race`, integração com Postgres e Keycloak reais, `govulncheck`, Sonar **bloqueante**; `api-gate` obrigatório | `.github/workflows/api-ci.yml`, `.github/rulesets/develop.json` |
| F-03 `web-ci` com gate | Biome, ESLint (StyleX), `tsc`, i18n, estrutura, Vitest, Playwright, Sonar e **Lighthouse** (mobile, login e cadastro); `web-gate` obrigatório | `.github/workflows/web-ci.yml`, `web/scripts/check-lighthouse.mjs` |
| F-04 Deploy automático em tst | push na `develop` → imagem → pacote de configuração e segredos → migração → `up`; **smoke test pós-deploy**; **rollback exercitado** | `deploy.yml`; rollback de 26/10/2026 em `docs/ci-cd.md` (20 s para voltar, 9 s para avançar) |
| F-05 Versão visível | `GET /api/v1/version` com versão, ambiente e commit; selo de versão no app; versão automática por componente | `VersionFooter`, [ADR 0030](../adr/0030-versao-automatica-por-componente-com-bump-na-develop.md) |
| F-06 Fluxo de release | **Substituído** pela versão automática por componente com bump na `develop` (o `release-please` foi removido). Promoção para prd fica no Sprint 6 (O-02) | ADR 0030 |
| I-01 Login | BFF OIDC com cookie `httpOnly` (`__Host-`), sair, sessões, passkeys, 2FA opcional, e-mail e senha (Argon2id), cadastro por e-mail, Google e GitHub | `docs/features/2026-10-13-login-bff.md` e seguintes |
| O-03 Segurança no CI | CodeQL, gitleaks, dependency-review, **Trivy nas imagens**, ações fixadas por SHA, configuração do Renovate | `.github/workflows/security.yml`, `.trivyignore`, `renovate.json` |

## Exceções registradas
1. **Renovate:** `renovate.json` está pronto, mas o **app do GitHub não está instalado** (nenhuma das PRs do repositório é de bot). Ação do dono: <https://github.com/apps/renovate> → Install → só o repositório `wordloop`. Não bloqueia nada, mas as atualizações automáticas de dependência só começam depois disso.
2. **Tela "Olá" e rodapé do roteiro** foram além do texto original: o rodapé virou um selo de versão em Configurações (a barra inferior ocupa o lugar do rodapé). A versão é por componente e automática, não o `v0.1.0` único do roteiro.

## Achados que ficam como dívida
- **CLS 0,39 na tela de login** (medido pelo Lighthouse; recomendado ≤ 0,1): a página "pula" enquanto carrega.
- **Exceções do Trivy com validade** em `.trivyignore` (`curl`, `libcurl`, `brotli-libs` da imagem base do web), até 27/11/2026.
- Segue sem **Trivy como check obrigatório** do ruleset (hoje informativo, em *Security* e na PR).
