FROM golang:1.25 AS base

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

FROM base AS healthcheck-build
WORKDIR /app

COPY healthcheck healthcheck

WORKDIR /app/healthcheck

RUN go build -a -installsuffix cgo -o /go/bin/healthcheck

FROM base AS dev
WORKDIR /app/finance-api

COPY --from=healthcheck-build /go/bin/healthcheck /go/bin/healthcheck

RUN go install github.com/air-verse/air@latest && go install github.com/go-delve/delve/cmd/dlv@latest

HEALTHCHECK --interval=5s --timeout=5s --start-period=5s --retries=5 CMD [ "/go/bin/healthcheck" ]
EXPOSE 8080
EXPOSE 2345

ENTRYPOINT ["air"]

FROM base AS app-build
WORKDIR /app

COPY go.mod go.sum ./

COPY apierror apierror
COPY finance-api finance-api
COPY shared shared

RUN go mod download

WORKDIR /app/finance-api

RUN go build -a -installsuffix cgo -o /go/bin/finance-api

FROM scratch

COPY --from=base /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=base /etc/passwd /etc/passwd
COPY --from=base /etc/group /etc/group

COPY --from=app-build /go/bin/finance-api finance-api
COPY --from=healthcheck-build /go/bin/healthcheck healthcheck

USER app-user:app-user

HEALTHCHECK --interval=5s --timeout=5s --start-period=5s --retries=5 CMD [ "/healthcheck" ]

ENTRYPOINT ["./finance-api"]
