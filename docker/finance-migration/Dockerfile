FROM golang:alpine3.23@sha256:18b460dd17542c2ba43299a633cf6ebfc1115101509531471d7cfce1019af083 AS build

RUN adduser \
    --disabled-password \
    --gecos "" \
    --home "/nonexistent" \
    --shell "/sbin/nologin" \
    --no-create-home \
    --uid 65532 \
    migration-user

RUN CGO_ENABLED=0 go install github.com/pressly/goose/v3/cmd/goose@v3.27.2

COPY migrations/ /database/
RUN rm /database/1_baseline.sql


FROM scratch
COPY --from=busybox:1.38.0-uclibc@sha256:eea4ff5612c911abd1d0e9ed47ba642547b01c3490877d9c1bb5fd6346462da4 /bin/sh /bin/sh
COPY --from=build /etc/passwd /etc/passwd
COPY --from=build /etc/group /etc/group
COPY --from=build /go/bin/goose /go/bin/goose
COPY --from=build /database/ /database/
COPY docker/finance-migration/migrate.sh /go/bin/migrate.sh

USER migration-user

ENTRYPOINT [ "/bin/sh", "/go/bin/migrate.sh" ]
