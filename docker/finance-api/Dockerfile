FROM golang:1.24 AS base
WORKDIR /app

ENV CGO_ENABLED=0 GOOS=linux

RUN update-ca-certificates

FROM base AS healthcheck-build
WORKDIR /app

COPY healthcheck healthcheck

WORKDIR /app/healthcheck

RUN go build -a -installsuffix cgo -o /go/bin/healthcheck

FROM base AS dev
WORKDIR /app/finance-api

COPY --from=healthcheck-build /go/bin/healthcheck healthcheck

RUN go install github.com/air-verse/air@latest && go install github.com/go-delve/delve/cmd/dlv@latest
EXPOSE 8080
EXPOSE 2345

HEALTHCHECK --interval=5s --timeout=5s --start-period=5s --retries=5 CMD [ "/app/finance-api/healthcheck" ]

ENTRYPOINT ["air"]

FROM base AS build-env
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

WORKDIR /app/finance-api

RUN go build -a -installsuffix cgo -o /go/bin/finance-api

FROM alpine:3
WORKDIR /go/bin

RUN apk --update --no-cache add \
    ca-certificates \
    tzdata

# Patch vulnerabilities
RUN apk upgrade --no-cache busybox libcrypto3 libssl3 musl musl-utils

COPY --from=build-env /go/bin/finance-api finance-api
ENTRYPOINT ["./finance-api"]
