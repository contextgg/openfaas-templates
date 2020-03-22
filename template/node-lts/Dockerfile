FROM openfaas/of-watchdog:0.7.4 as watchdog
FROM node:lts-alpine

RUN mkdir -p /app

WORKDIR /app

ENV NODE_ENV=production
RUN npm install pm2@latest -g

COPY function/package.json ./
COPY function/package*.json ./
RUN npm install

COPY function/ ./
RUN npm run build

COPY --from=watchdog /fwatchdog /usr/bin/fwatchdog
RUN chmod +x /usr/bin/fwatchdog

ENV cgi_headers="true"
ENV fprocess="npm run start"
ENV mode="http"
ENV upstream_url="http://127.0.0.1:5000"

ENV exec_timeout="10s"
ENV write_timeout="15s"
ENV read_timeout="15s"

HEALTHCHECK --interval=3s CMD [ -e /tmp/.lock ] || exit 1

CMD ["fwatchdog"]