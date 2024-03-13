FROM postgres:12 as db
WORKDIR /app
COPY ./sirius-db/init.sh /docker-entrypoint-initdb.d
COPY ./sirius-db/dump.sql ./scripts/db/dump.sql
