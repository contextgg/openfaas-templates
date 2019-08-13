FROM node:12.8.0-alpine as builder

RUN mkdir -p /home/app

WORKDIR /home/app

COPY function/package.json ./
COPY function/yarn.lock ./
RUN yarn install

# COPY function files and folders
COPY function/ ./

RUN yarn build

FROM openfaas/of-watchdog:0.5.3 as watchdog
FROM node:12.8.0-alpine as ship

RUN apk --no-cache add curl ca-certificates \
    && addgroup -S app && adduser -S -g app app

# Turn down the verbosity to default level.
ENV NPM_CONFIG_LOGLEVEL warn

RUN mkdir -p /home/app

# Wrapper/boot-strapper
WORKDIR /home/app

# Add the serve
RUN yarn global add serve

COPY --from=builder /home/app/build /home/app/build
COPY --from=watchdog /fwatchdog /usr/bin/fwatchdog
RUN chmod +x /usr/bin/fwatchdog

# chmod for tmp is for a buildkit issue (@alexellis)
RUN chown app:app -R /home/app \
    && chmod 777 /tmp

USER app

ENV cgi_headers="true"
ENV fprocess="serve -s build"
ENV mode="http"
ENV upstream_url="http://127.0.0.1:5000"

ENV exec_timeout="10s"
ENV write_timeout="15s"
ENV read_timeout="15s"

HEALTHCHECK --interval=3s CMD [ -e /tmp/.lock ] || exit 1

CMD ["fwatchdog"]