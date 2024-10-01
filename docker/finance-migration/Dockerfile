FROM alpine AS build

RUN apk add --no-cache curl

RUN curl -fsSL \
    https://raw.githubusercontent.com/pressly/goose/master/install.sh |\
    sh
COPY migrations/ /database/
RUN rm /database/1_baseline.sql


FROM scratch
COPY --from=busybox:1.36.1-uclibc /bin/sh /bin/sh
COPY --from=build /usr/local/bin/goose /usr/local/bin/goose
COPY --from=build /database/ /database/
COPY docker/finance-migration/migrate.sh /usr/local/bin/migrate.sh

ENTRYPOINT [ "/bin/sh", "/usr/local/bin/migrate.sh" ]
