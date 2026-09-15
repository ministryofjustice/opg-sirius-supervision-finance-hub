FROM golang:1.26-alpine@sha256:ce864e7223ac17b1775e6fd0b4c0db580c2eb50e7953a427916379e4b92a1628

WORKDIR /app

COPY go.mod go.sum ./
COPY api-mocks/holidays-api .
RUN go mod download

RUN go build -a -installsuffix cgo -o holidays-api

ENTRYPOINT ["./holidays-api"]
