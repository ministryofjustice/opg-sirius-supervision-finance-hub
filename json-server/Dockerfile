## build json-server for mock API
FROM node:24-alpine3.23@sha256:9ec4a2e289874ed0d722e1772ec2de45d2801541db8612f3638b26f128c69ac2 AS json-server

WORKDIR /app

EXPOSE 3000

COPY json-server .

RUN apk add --no-cache git
RUN npm ci --ignore-scripts

ENTRYPOINT ["npm", "run", "serve"]