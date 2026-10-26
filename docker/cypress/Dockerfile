FROM cypress/included:16.2.0@sha256:c196ed896aec24c7ea8fc1d4185b210493e311339b6c09d0153f9b8dfc6a6058

WORKDIR /root

COPY cypress/package.json .
COPY cypress/package-lock.json .

RUN apt-get update && apt-get install -y --no-install-recommends git && rm -rf /var/lib/apt/lists/*
RUN npm ci --ignore-scripts

ENV CYPRESS_VIDEO=false
ENV CYPRESS_baseUrl=http://finance-hub:8888/finance

COPY cypress .
