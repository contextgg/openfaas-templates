# https://hub.docker.com/_/microsoft-dotnet-core
FROM mcr.microsoft.com/dotnet/core/sdk:3.1 AS build
WORKDIR /source

# copy csproj and restore as distinct layers
COPY function/*.csproj ./
RUN dotnet restore -r linux-musl-x64

# copy everything else and build app
COPY function/. ./
RUN dotnet publish -c release -o /app -r linux-musl-x64 --self-contained false --no-restore

FROM openfaas/of-watchdog:0.7.2 as watchdog
FROM mcr.microsoft.com/dotnet/core/aspnet:3.1-alpine

# Add non root user and certs
RUN apk --no-cache add ca-certificates \
    && addgroup -S app && adduser -S -g app app \
    && mkdir -p /home/app \
    && chown app /home/app

WORKDIR /home/app

COPY --from=build       /app            .
COPY --from=watchdog    /fwatchdog      .

RUN chmod +x ./fwatchdog
RUN chmod +x ./function
RUN chown -R app /home/app

USER app

ENV ASPNETCORE_URLS http://*:8082
ENV fprocess="./function"
ENV mode="http"
ENV upstream_url="http://127.0.0.1:8082"

CMD ["./fwatchdog"]