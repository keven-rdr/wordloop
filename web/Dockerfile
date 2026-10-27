# syntax=docker/dockerfile:1
# Imagem do web: SEM build-arg de ambiente. A configuracao de cada ambiente entra em runtime (environment.json montado).
# Contexto de build: a RAIZ do repositorio (o cliente e gerado de api/openapi). Uso:
#   docker build -f web/Dockerfile -t wordloop-web --build-arg VERSION=0.1.0 --build-arg COMMIT=abc1234 .
FROM --platform=$BUILDPLATFORM node:24-alpine AS build
ARG VERSION=dev
ARG COMMIT=unknown
WORKDIR /repo
COPY web/package.json web/package-lock.json web/
RUN cd web && npm ci
COPY api/openapi api/openapi
COPY web web
RUN cd web && npm run gen && VERSION=$VERSION COMMIT=$COMMIT npm run build

# O estagio final so COPIA arquivos estaticos: o mesmo build serve amd64 e arm64.
# 1.30 e a linha estavel mantida do nginx; a tag 1.28-alpine foi abandonada (ultima imagem de fev/2026, 72 CVEs altas
# ou criticas ja corrigidas no Alpine). Atualize a linha quando o nginx publicar a proxima estavel.
FROM nginxinc/nginx-unprivileged:1.30-alpine
COPY web/nginx.conf /etc/nginx/conf.d/default.conf
COPY --from=build /repo/web/dist /usr/share/nginx/html
EXPOSE 8080
