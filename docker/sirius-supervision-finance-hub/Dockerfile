FROM golang:1.21 as base
WORKDIR /app

ENV CGO_ENABLED=0 GOOS=linux

RUN update-ca-certificates

FROM node:20.2.0-alpine3.16 as asset-env

WORKDIR /app

RUN mkdir -p web/static

COPY web/assets web/assets
COPY package.json .
COPY yarn.lock .

RUN yarn install
RUN yarn build

FROM base AS dev
WORKDIR /app

COPY --from=asset-env /app/web/static web/static

FROM base as build-env

WORKDIR /app

COPY go.mod .
COPY go.sum .

RUN go mod download

COPY . .

RUN go build -a -installsuffix cgo -o /go/bin/opg-sirius-supervision-finance-hub

FROM alpine:3

WORKDIR /go/bin

RUN apk --update --no-cache add \
    ca-certificates \
    tzdata

# Patch vulnerabilities
RUN apk upgrade --no-cache libcrypto3 libssl3

COPY --from=build-env /go/bin/opg-sirius-supervision-finance-hub opg-sirius-supervision-finance-hub
COPY --from=build-env /app/web/template web/template
COPY --from=asset-env /app/web/static web/static
ENTRYPOINT ["./opg-sirius-supervision-finance-hub"]
