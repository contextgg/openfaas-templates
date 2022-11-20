ARG FUNCTION_IMG
ARG FROM=golang:1.19

FROM scratch as server
ARG FILES=golang
COPY ${FILES} /app

FROM ${FUNCTION_IMG} as function
FROM ${FROM} as build

ARG DB_HOSTNAME
ENV DB_HOSTNAME=$DB_HOSTNAME

RUN mkdir /app

COPY --from=server /app /app
RUN rm -rf /app/function
COPY --from=function /app /app/function

WORKDIR /app/function
RUN test -z "$(gofmt -l $(find . -type f -name '*.go' -not -path "./vendor/*"))" || { echo "Run \"gofmt -s -w\" on your Golang code"; exit 1; }
RUN GOOS=linux go test -p=1 ./... -cover


WORKDIR /app/handler
RUN CGO_ENABLED=0 GOOS=linux \
    go build --ldflags "-s -w" -a -installsuffix cgo -o app .

FROM alpine:3.16

# Add non root user and certs
RUN apk --no-cache add ca-certificates \
    && addgroup -S app && adduser -S -g app app
RUN mkdir -p /home/app \
    && chown app /home/app
WORKDIR /home/app

COPY --from=build --chown=app /app/handler/app .

USER app

EXPOSE 8080
EXPOSE 8081
EXPOSE 8082

CMD ["./app"]