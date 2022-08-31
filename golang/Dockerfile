ARG FROM=golang:1.19
FROM $FROM as build

ENV GOPATH /app
ENV GO111MODULE on

ARG DB_HOSTNAME
ENV DB_HOSTNAME=$DB_HOSTNAME

RUN mkdir -p /app/src
WORKDIR /app/src

COPY . .

# Run a gofmt and exclude all vendored code.
RUN test -z "$(gofmt -l $(find . -type f -name '*.go' -not -path "./handler/*" -not -path "./function/vendor/*"))" || { echo "Run \"gofmt -s -w\" on your Golang code"; exit 1; }

RUN GOOS=linux go test -p=1 handler/... -cover
RUN CGO_ENABLED=0 GOOS=linux go build --ldflags "-s -w" -a -installsuffix cgo -o app handler

FROM alpine:3.16

# Add non root user and certs
RUN apk --no-cache add ca-certificates \
    && addgroup -S app && adduser -S -g app app
RUN mkdir -p /home/app \
    && chown app /home/app

WORKDIR /home/app

COPY --from=build --chown=app /app/src/app .

USER app

EXPOSE 8080
EXPOSE 8081
EXPOSE 8082

CMD ["./app"]