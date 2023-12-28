# syntax = docker/dockerfile:1.2

# Base stage
FROM golang:1.21-alpine AS base
WORKDIR /payroll
COPY . .
RUN --mount=type=cache,target=/go/pkg/mod \
  go mod download

# Build stage
FROM base AS build
ARG TARGETOS
ARG TARGETARCH
ARG VERSION
ARG BINARY_NAME
RUN --mount=type=cache,target=/root/.cache/go-build \
  GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build -o ./build/linux-amd64-${VERSION}/${BINARY_NAME} .

# Unit test stage
FROM base AS unit-test
RUN apk add --no-cache build-base
RUN --mount=target=. \
  --mount=type=cache,target=/root/.cache/go-build \
   go test ./...

# Integration test stage
# FROM base AS intergration-test
# RUN --mount=target=. \
#   --mount=type=cache,target=/root/.cache/go-build \
#    go test -v -p 1 --race -timeout 20m ./...

FROM golangci/golangci-lint:v1.49.0 AS lint-base

# Lint check stage
FROM base AS lint
RUN --mount=target=. \
    --mount=from=lint-base,src=/usr/bin/golangci-lint,target=/usr/bin/golangci-lint \
    --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    --mount=type=cache,target=/root/.cache/golangci-lint \
    golangci-lint run ./...


# Run stage
FROM alpine:latest as dev
ARG VERSION
ARG BINARY_NAME
RUN apk add --no-cache bash curl ca-certificates tzdata
COPY --from=build /payroll/build/linux-amd64-${VERSION}/${BINARY_NAME} /
COPY --from=build /payroll/config/.payroll.yaml /root/
EXPOSE 8088
CMD [ "/payroll" ]