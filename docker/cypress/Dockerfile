FROM cypress/included:15.18.1@sha256:9d3e6ac7d4a725e802080fb45f4c8d08d1a0eb78b3907232ffb3a42cd1f05e33

WORKDIR /root

COPY cypress/package.json .
COPY cypress/package-lock.json .

RUN apt-get update && apt-get install -y --no-install-recommends git && rm -rf /var/lib/apt/lists/*
RUN npm ci --ignore-scripts

ENV CYPRESS_VIDEO=false
ENV CYPRESS_baseUrl=http://finance-hub:8888/finance

COPY cypress .
