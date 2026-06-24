FROM golang:1.26-bookworm@sha256:5ae05f331ade3e10c2647be439a4f42dae789002f7d2e2112dcbf3d26bec25e9

RUN go install gotest.tools/gotestsum@latest

RUN go install github.com/pact-foundation/pact-go/v2@v2.4.2
RUN pact-go -l DEBUG install
