FROM golang:1.26-alpine@sha256:3ad57304ad93bbec8548a0437ad9e06a455660655d9af011d58b993f6f615648

WORKDIR /app

COPY go.mod go.sum ./
COPY api-mocks/holidays-api .
RUN go mod download

RUN go build -a -installsuffix cgo -o holidays-api

ENTRYPOINT ["./holidays-api"]
