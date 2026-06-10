FROM cypress/included:15.16.0@sha256:057f6c6bbb661342ea712f94f5b870823823e3617354db7a78a01f2aabbfc2ad

WORKDIR /root

COPY cypress/package.json .
COPY cypress/yarn.lock .
RUN yarn install --frozen-lockfile --ignore-scripts

ENV CYPRESS_VIDEO=false
ENV CYPRESS_baseUrl=http://finance-hub:8888/finance

COPY cypress .
