## build json-server for mock API
FROM node:24-alpine3.23@sha256:2bdb65ed1dab192432bc31c95f94155ca5ad7fc1392fb7eb7526ab682fa5bf14 AS json-server

WORKDIR /app

EXPOSE 3000

COPY json-server .

RUN yarn install --frozen-lockfile --ignore-scripts

ENTRYPOINT ["yarn", "serve"]