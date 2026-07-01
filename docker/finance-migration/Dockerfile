FROM alpine:3.23.4@sha256:5b10f432ef3da1b8d4c7eb6c487f2f5a8f096bc91145e68878dd4a5019afde11 AS build

RUN apk add --no-cache curl

RUN curl -fsSL \
    https://raw.githubusercontent.com/pressly/goose/master/install.sh |\
    sh

COPY migrations/ /database/
RUN rm /database/1_baseline.sql


FROM scratch
COPY --from=busybox:1.38.0-uclibc@sha256:eea4ff5612c911abd1d0e9ed47ba642547b01c3490877d9c1bb5fd6346462da4 /bin/sh /bin/sh
COPY --from=build /usr/local/bin/goose /usr/local/bin/goose
COPY --from=build /database/ /database/
COPY docker/finance-migration/migrate.sh /usr/local/bin/migrate.sh

ENTRYPOINT [ "/bin/sh", "/usr/local/bin/migrate.sh" ]
