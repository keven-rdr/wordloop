# syntax=docker/dockerfile:1
# Imagem da API. Multi-arquitetura SEM emulacao: o Go cross-compila no runner (BUILDPLATFORM) para o destino (TARGETARCH).
# Contexto de build: a pasta api/. Uso: docker build -t wordloop-api --build-arg VERSION=0.1.0 --build-arg COMMIT=abc1234 api
FROM --platform=$BUILDPLATFORM golang:1.27-alpine AS build
ARG TARGETOS
ARG TARGETARCH
ARG VERSION=dev
ARG COMMIT=unknown
WORKDIR /src
COPY go.mod go.sum* ./
RUN go mod download
COPY . .
# O codigo gerado do contrato nao e versionado: gera dentro da imagem, a partir do commit (hermetico).
RUN go generate ./... \
 && CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH go build -trimpath \
    -ldflags "-s -w \
      -X github.com/keven-rdr/wordloop/api/internal/platform/version.Version=${VERSION} \
      -X github.com/keven-rdr/wordloop/api/internal/platform/version.Commit=${COMMIT}" \
    -o /out/api ./cmd/api

FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=build /out/api /api
EXPOSE 8080
USER nonroot
ENTRYPOINT ["/api"]
