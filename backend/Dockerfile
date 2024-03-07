# Use the official Golang image as the base image
FROM golang:1.22-alpine AS builder

# Set the current working directory inside the container
WORKDIR /app

# Copy the Go module files
COPY go.mod go.sum ./

# Download and install Go dependencies
RUN go mod download

# Copy the rest of the application source code
COPY . .

# Build the Go application
WORKDIR /app/backend

RUN go build -a -installsuffix cgo -o /go/bin/opg-sirius-finance-api

FROM alpine:3

WORKDIR /go/bin

RUN apk --update --no-cache add \
    ca-certificates \
    tzdata

# Patch vulnerabilities
RUN apk upgrade --no-cache libcrypto3 libssl3

COPY --from=builder /go/bin/opg-sirius-finance-api opg-sirius-finance-api
ENTRYPOINT ["./opg-sirius-finance-api"]
