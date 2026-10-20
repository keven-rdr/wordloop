# ADR 0030 — Versão automática por componente, com bump de patch gravado na `develop`

**Status:** aceita (2026-10-20). Substitui o [ADR 0018](0018-release-please-por-componente-build-uma-vez-e.md) na parte de versionamento.

## Contexto
Para saber se dev, tst e prd estão atualizados, a versão tem de subir sozinha a cada atualização e ser **um número real por componente** (web e API), sem depender de uma pessoa mergear uma PR de release. O release-please só sobe a versão quando a PR de release é mergeada, e entre releases o app mostrava um número derivado de tag (`0.4.0-5-gabc`) ou um valor sem sentido.

## Decisão
Cada componente tem `api/VERSION` e `web/VERSION` (`MAJOR.MINOR.PATCH`). A cada merge na `develop` que muda o componente, o job `image` roda `scripts/bump-version.sh`: sobe o **patch**, grava o número num commit `chore(release): ... [skip ci]` na `develop` e usa o mesmo número como `VERSION` da imagem. Minor e major mudam editando o arquivo na PR. O release-please foi removido.

O commit sai de um worktree separado (a imagem é construída do commit do merge) e repete até 5 vezes se outro bump gravar antes. Como a `develop` exige PR, o push usa o secret `RELEASE_PUSH_TOKEN` (de um admin) e o ruleset dá bypass ao papel *Repository admin*. A variável `VERSION_BUMP_ENABLED=true` liga o bump; desligada, o CI usa o número que está no arquivo.

## Consequências
Uma versão por componente, que só sobe quando ele muda. Há um commit de bot por merge e uma exceção controlada à regra de não comitar direto na `develop`. Sem tags nem changelog automático. A promoção para prd continua por re-etiqueta da imagem `sha-<commit>`.

## Alternativas descartadas
release-please (a versão só sobe com PR de release mergeada); contagem de merges sem commit (número sintético, igual para web e API); versão derivada de `git describe` (sem tag não há número real).

## Fonte
`docs/ci-cd.md`, seção *Versão do app*.
