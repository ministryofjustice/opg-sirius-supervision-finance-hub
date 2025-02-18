FROM golang:1.23-alpine AS base
WORKDIR /app

ENV CGO_ENABLED=0 GOOS=linux

RUN update-ca-certificates

FROM node:22-alpine3.19 AS asset-env
WORKDIR /app

RUN mkdir -p web/static

COPY finance-hub/web/assets web/assets
COPY finance-hub/package.json .
COPY finance-hub/yarn.lock .

RUN yarn install
RUN yarn build

FROM base AS dev
WORKDIR /app/finance-hub

COPY --from=asset-env /app/web/static web/static

RUN go install github.com/cosmtrek/air@v1.52.1 && go install github.com/go-delve/delve/cmd/dlv@latest
EXPOSE 8080
EXPOSE 2345

ENTRYPOINT ["air"]

FROM base AS build-env
WORKDIR /app

COPY go.mod .
COPY go.sum .

RUN go mod download

COPY . .

WORKDIR /app/finance-hub

RUN go build -a -installsuffix cgo -o /go/bin/finance-hub

FROM alpine:3
WORKDIR /go/bin

RUN apk --update --no-cache add \
    ca-certificates \
    tzdata

# Patch vulnerabilities
RUN apk upgrade --no-cache busybox libcrypto3 libssl3 musl musl-utils

COPY --from=build-env /go/bin/finance-hub finance-hub
COPY --from=build-env /app/finance-hub/web/template web/template
COPY --from=asset-env /app/web/static web/static
ENTRYPOINT ["./finance-hub"]
