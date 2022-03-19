ARG FROM=golang:1.17
FROM $FROM as build

ENV GOPATH /app
ENV GO111MODULE auto

ARG DB_HOSTNAME
ENV DB_HOSTNAME=$DB_HOSTNAME

RUN mkdir -p /app/src
WORKDIR /app/src

COPY handler ./handler
COPY function ./function

# Run a gofmt and exclude all vendored code.
RUN test -z "$(gofmt -l $(find . -type f -name '*.go' -not -path "./handler/*" -not -path "./function/vendor/*"))" || { echo "Run \"gofmt -s -w\" on your Golang code"; exit 1; }

ARG LDFLAGS="-linkmode=external -w -s -extldflags '-static'"
RUN echo GOOS=linux go build -ldflags=\"${LDFLAGS}\" -o=handler handler > build.sh && chmod +x ./build.sh && ./build.sh
RUN go test -p=1 ./... -cover

FROM alpine:3.10

# Add non root user and certs
RUN apk --no-cache add ca-certificates \
    && addgroup -S app && adduser -S -g app app \
    && mkdir -p /home/app \
    && chown app /home/app

WORKDIR /home/app

COPY --from=build /app/src/handler/handler    .

RUN chown -R app /home/app

USER app

EXPOSE 8080
EXPOSE 8081
EXPOSE 8082

CMD ["./handler"]