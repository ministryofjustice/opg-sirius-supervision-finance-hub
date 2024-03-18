FROM postgres:12 as db
WORKDIR /app
COPY ./sirius-db/init.sh /docker-entrypoint-initdb.d
