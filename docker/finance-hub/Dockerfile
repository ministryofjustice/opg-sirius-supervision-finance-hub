FROM golang:1.26.10@sha256:672c944cfd07599b5f1bb179869270cdc94c6254e41db4d431259693cb286b82 AS base

RUN adduser \
    --disabled-password \
    --gecos "" \
    --home "/home/app-user" \
    --shell "/sbin/nologin" \
    --uid 65532 \
    app-user

WORKDIR /app

ENV CGO_ENABLED=0 GOOS=linux

RUN update-ca-certificates

FROM node:24-alpine3.24@sha256:06d7adf3e0ef61a13f1d28682726366f97a15aa1888c934d3ecec86c33918b53 AS asset-env
WORKDIR /app

RUN mkdir -p web/static

COPY finance-hub/web/assets web/assets
COPY finance-hub/package*.json .
COPY finance-hub/.npmrc .

RUN apk add --no-cache git
RUN npm ci --ignore-scripts
RUN npm run build

FROM base AS healthcheck-build
WORKDIR /app

COPY healthcheck healthcheck

WORKDIR /app/healthcheck

RUN go build -a -installsuffix cgo -o /go/bin/healthcheck

FROM base AS dev
WORKDIR /app/finance-hub

COPY --from=asset-env /app/web/static web/static
COPY --from=healthcheck-build /go/bin/healthcheck /go/bin/healthcheck

RUN go install github.com/air-verse/air@latest && go install github.com/go-delve/delve/cmd/dlv@latest
RUN mkdir -p /app/finance-hub/tmp && chown -R app-user:app-user /app/finance-hub
ENV HOME=/home/app-user
ENV GOCACHE=/tmp/.cache/go-build
ENV GOMODCACHE=/tmp/.cache/go-mod

HEALTHCHECK --interval=5s --timeout=5s --start-period=5s --retries=5 CMD [ "/go/bin/healthcheck" ]
EXPOSE 8080
EXPOSE 2345

ENTRYPOINT ["air"]

FROM base AS app-build
WORKDIR /app

COPY go.mod .
COPY go.sum .

COPY apierror apierror
COPY finance-hub finance-hub
COPY shared shared

RUN go mod download

WORKDIR /app/finance-hub

RUN go build -a -installsuffix cgo -o /go/bin/finance-hub

FROM scratch

COPY --from=base /usr/share/zoneinfo /usr/share/zoneinfo
COPY --from=base /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=base /etc/passwd /etc/passwd
COPY --from=base /etc/group /etc/group

COPY --from=app-build /go/bin/finance-hub finance-hub
COPY --from=app-build /app/finance-hub/web/template web/template
COPY --from=healthcheck-build /go/bin/healthcheck healthcheck
COPY --from=asset-env /app/web/static web/static

USER app-user:app-user

HEALTHCHECK --interval=5s --timeout=5s --start-period=5s --retries=5 CMD [ "/healthcheck" ]

ENTRYPOINT ["./finance-hub"]
