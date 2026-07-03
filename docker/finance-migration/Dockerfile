FROM golang:alpine3.23@sha256:d94c1169c47217d91242d8ae44891081169e1e99017804f94c3d9309b039d2de AS build

RUN CGO_ENABLED=0 go install github.com/pressly/goose/v3/cmd/goose@v3.27.2

COPY migrations/ /database/
RUN rm /database/1_baseline.sql


FROM scratch
COPY --from=busybox:1.38.0-uclibc@sha256:eea4ff5612c911abd1d0e9ed47ba642547b01c3490877d9c1bb5fd6346462da4 /bin/sh /bin/sh
COPY --from=build /usr/local/bin/goose /usr/local/bin/goose
COPY --from=build /database/ /database/
COPY docker/finance-migration/migrate.sh /usr/local/bin/migrate.sh

ENTRYPOINT [ "/bin/sh", "/usr/local/bin/migrate.sh" ]
