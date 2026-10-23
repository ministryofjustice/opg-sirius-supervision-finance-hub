FROM golang:1.26.10-alpine@sha256:27f2d3b83cfc5803c8078f8e730af0f7c431bf067afaec8f60a36fb763187cba

WORKDIR /app

COPY go.mod go.sum ./
COPY api-mocks/holidays-api .
RUN go mod download

RUN go build -a -installsuffix cgo -o holidays-api

ENTRYPOINT ["./holidays-api"]
