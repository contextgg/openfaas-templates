FROM node:lts-alpine as builder

RUN mkdir -p /app

WORKDIR /app

COPY function/package.json ./
COPY function/package*.json ./
RUN npm install

# COPY function files and folders
COPY function/ ./

RUN npm run build

FROM openfaas/of-watchdog:0.7.4 as watchdog
FROM nginx:stable

COPY nginx.conf /etc/nginx/conf.d/default.conf
RUN echo "daemon off;" >> /etc/nginx/nginx.conf

RUN mkdir -p /app

WORKDIR /app

COPY --from=builder /app/build /app/build
COPY --from=watchdog /fwatchdog /usr/bin/fwatchdog
RUN chmod +x /usr/bin/fwatchdog

ENV cgi_headers="true"
ENV fprocess="nginx"
ENV mode="http"
ENV upstream_url="http://127.0.0.1:5000"

ENV exec_timeout="10s"
ENV write_timeout="15s"
ENV read_timeout="15s"

HEALTHCHECK --interval=3s CMD [ -e /tmp/.lock ] || exit 1

CMD ["fwatchdog"]