# wordloop — documentação do planejamento

PWA de vocabulário de inglês com repetição espaçada (Go + React + StyleX). Planejamento em 6 fases, concluído em 2026-10-06. **Comece por [fase-6/06-roadmap-backlog.md](fase-6/06-roadmap-backlog.md).**

| Fase | Documento | O que traz |
|---|---|---|
| 0 | [fase-0 — alinhamento](00-fase-0-alinhamento.md) · [VM e domínio](00b-vm-dominio.md) | decisões iniciais, VM Contabo, domínio `wordloop` |
| 1 | [SRS](fase-1/01-srs.md) · [script de referência](fase-1/srs_reference.py) · [exemplo](fase-1/worked_example.py) | evidência, agendador, `simple_v1` × `fsrs6`, log, exercícios |
| 2 | [pedagogia e conteúdo](fase-2/02-pedagogia-conteudo.md) · [DDL de conteúdo](fase-2/content-schema.sql) | trilha, cobertura, imagens, licenças, modelo de conteúdo |
| 3 | [arquitetura e features](fase-3/03-arquitetura-features.md) · [pastas](fase-3/03b-estrutura-pastas.md) · [diagramas](fase-3/03c-diagramas.md) · [design system](fase-3/03d-design-system-stylex.md) · [DDL do usuário](fase-3/schema-learning.sql) · [OpenAPI](fase-3/openapi-sketch.yaml) | identidade, push, API, front, dados |
| 4 | [engenharia](fase-4/04-engenharia.md) · [workflows](fase-4/workflows/) · [deploy](fase-4/deploy/deploy-skeleton.md) · [kit de IA](fase-4/kit-ia/) | repositório, CI/CD, versão, ambientes, qualidade |
| 5 | [UX/UI](fase-5/05-ux-ui-fluxos.md) · [wireframes](fase-5/05b-wireframes.md) | telas, fluxos, acessibilidade, PWA |
| 6 | [roadmap e backlog](fase-6/06-roadmap-backlog.md) · [seed](fase-6/seed/README.md) | sprints, MoSCoW, seed e amostra |
| 0a | [relatório dos spikes](sprint-0a/spike-report.md) · [web](sprint-0a/web-stack/) · [contrato](sprint-0a/contract/) · [Go](sprint-0a/go-stack/) · [PostgreSQL](sprint-0a/postgres/) | resultado real da pilha (front, contrato, Go, PostgreSQL); Sprint 0a concluído |
| — | [ADRs](adr/README.md) · [referências auditadas](referencias.md) | 27 decisões registradas; nível de verificação de cada fonte |

**Convenção de rótulos:** `[PREMISSA]` assumi · `[VERIFICAR]` não confirmei em fonte primária · `[DECISÃO SUA]` só o dono decide. Esqueletos de workflow, Compose, SQL e Mermaid **não foram executados** (exceto onde o documento diz o contrário).
