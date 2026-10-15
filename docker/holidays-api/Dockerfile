FROM golang:1.26.6-alpine@sha256:3889b425f035be855a72fb4755265311293b6d414521f0a519d819df32222d83

WORKDIR /app

COPY go.mod go.sum ./
COPY api-mocks/holidays-api .
RUN go mod download

RUN go build -a -installsuffix cgo -o holidays-api

ENTRYPOINT ["./holidays-api"]
