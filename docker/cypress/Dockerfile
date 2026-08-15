FROM cypress/included:15.20.0@sha256:5baa6e2f1831e7b9823246d22b5914fe77187a9bab030aa2f44fceabde17c9d1

WORKDIR /root

COPY cypress/package.json .
COPY cypress/package-lock.json .

RUN apt-get update && apt-get install -y --no-install-recommends git && rm -rf /var/lib/apt/lists/*
RUN npm ci --ignore-scripts

ENV CYPRESS_VIDEO=false
ENV CYPRESS_baseUrl=http://finance-hub:8888/finance

COPY cypress .
