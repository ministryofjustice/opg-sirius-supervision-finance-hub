FROM cypress/included:16.1.1@sha256:e652cda84399aaf99c68d1d0322085dc3faa7ca9ace88d9bb4f0f045d2ce6764

WORKDIR /root

COPY cypress/package.json .
COPY cypress/package-lock.json .

RUN apt-get update && apt-get install -y --no-install-recommends git && rm -rf /var/lib/apt/lists/*
RUN npm ci --ignore-scripts

ENV CYPRESS_VIDEO=false
ENV CYPRESS_baseUrl=http://finance-hub:8888/finance

COPY cypress .
