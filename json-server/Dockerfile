## build json-server for mock API
FROM node:24-alpine3.23@sha256:244cc2b53f46f9e876304391d17682b0ddae9ac33491f4857e25e35a36ba7995 AS json-server

WORKDIR /app

EXPOSE 3000

COPY json-server .

RUN apk add --no-cache git
RUN npm ci --ignore-scripts

ENTRYPOINT ["npm", "run", "serve"]