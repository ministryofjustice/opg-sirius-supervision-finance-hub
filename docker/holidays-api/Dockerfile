FROM golang:1.24-alpine

WORKDIR /app

COPY go.mod go.sum ./
COPY api-mocks/holidays-api .
RUN go mod download

RUN go build -a -installsuffix cgo -o holidays-api

ENTRYPOINT ["./holidays-api"]