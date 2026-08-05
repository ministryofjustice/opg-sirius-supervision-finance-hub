FROM cypress/included:15.18.0@sha256:dd99f59c8ef766157ce779222708007e9e7107980552dafdc6f037dd6a382524

WORKDIR /root

COPY cypress/package.json .
COPY cypress/package-lock.json .

RUN apt-get update \
    && apt-get install -y --no-install-recommends git \
    && rm -rf /var/lib/apt/lists/*
RUN npm ci --ignore-scripts

ENV CYPRESS_VIDEO=false
ENV CYPRESS_baseUrl=http://finance-hub:8888/finance

COPY cypress .
