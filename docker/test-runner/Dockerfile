FROM golang:1.26-bookworm@sha256:d1af0fd434fced72b3b4335440df9d0fd43a4c737ea14aed18b6b3f4e9aab58d

RUN go install gotest.tools/gotestsum@latest

RUN go install github.com/pact-foundation/pact-go/v2@v2.4.2
RUN pact-go -l DEBUG install
