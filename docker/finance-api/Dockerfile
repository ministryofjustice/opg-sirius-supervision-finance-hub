FROM golang:1.24.0-alpine AS base

RUN adduser \
    --disabled-password \
    --gecos "" \
    --home "/nonexistent" \
    --shell "/sbin/nologin" \
    --no-create-home \
    --uid 65532 \
    app-user

WORKDIR /app

ENV CGO_ENABLED=0 GOOS=linux

RUN update-ca-certificates

FROM base AS dev
WORKDIR /app/finance-api

RUN go install github.com/air-verse/air@latest && go install github.com/go-delve/delve/cmd/dlv@latest
EXPOSE 8080
EXPOSE 2345

ENTRYPOINT ["air"]

FROM base AS app-build
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

WORKDIR /app/finance-api

RUN go build -a -installsuffix cgo -o /go/bin/finance-api

FROM base AS healthcheck-build
WORKDIR /app

COPY healthcheck healthcheck

WORKDIR /app/healthcheck

RUN go build -a -installsuffix cgo -o /go/bin/healthcheck

FROM alpine:3
RUN apk --update --no-cache add \
    ca-certificates \
    tzdata

# Patch vulnerabilities
RUN apk upgrade --no-cache busybox libcrypto3 libssl3 musl musl-utils

COPY --from=base /etc/passwd /etc/passwd
COPY --from=base /etc/group /etc/group

RUN mkdir /app &&\
    chown -R app-user:app-user /app
COPY --chown=app-user:app-user --chmod=755 --from=app-build /go/bin/finance-api /app/finance-api
COPY --chown=app-user:app-user --chmod=755 --from=healthcheck-build /go/bin/healthcheck /app/healthcheck

USER app-user:app-user

HEALTHCHECK --interval=5s --timeout=5s --start-period=5s --retries=5 CMD [ "/app/healthcheck" ]

ENTRYPOINT ["/app/finance-api"]
