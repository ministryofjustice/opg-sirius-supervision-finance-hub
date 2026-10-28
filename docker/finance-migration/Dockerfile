FROM golang:1.26.10-alpine@sha256:27f2d3b83cfc5803c8078f8e730af0f7c431bf067afaec8f60a36fb763187cba AS build

RUN adduser \
    --disabled-password \
    --gecos "" \
    --home "/nonexistent" \
    --shell "/sbin/nologin" \
    --no-create-home \
    --uid 65532 \
    migration-user

RUN mkdir /tmp/goose && cd /tmp/goose && \
    go mod init goose-tool && \
    go get github.com/pressly/goose/v3/cmd/goose@v3.29.0 && \
    go get github.com/klauspost/compress@v1.20.2 && \
    CGO_ENABLED=0 go install github.com/pressly/goose/v3/cmd/goose

COPY migrations/ /database/
RUN rm /database/1_baseline.sql


FROM scratch
COPY --from=busybox:1.38.0-uclibc@sha256:b699b0c739b2d039afd8e2488a3ccfc0b9a86301d87633448d4a93c57bee8c76 /bin/sh /bin/sh
COPY --from=build /etc/passwd /etc/passwd
COPY --from=build /etc/group /etc/group
COPY --from=build /go/bin/goose /go/bin/goose
COPY --from=build /database/ /database/
COPY docker/finance-migration/migrate.sh /go/bin/migrate.sh

USER migration-user

ENTRYPOINT [ "/bin/sh", "/go/bin/migrate.sh" ]
