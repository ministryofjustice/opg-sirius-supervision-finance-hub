FROM cypress/included:14.5.1

WORKDIR /root

COPY cypress/package.json .
COPY cypress/yarn.lock .
RUN yarn install

ENV CYPRESS_VIDEO=false
ENV CYPRESS_baseUrl=http://finance-hub:8888/finance

COPY cypress .
