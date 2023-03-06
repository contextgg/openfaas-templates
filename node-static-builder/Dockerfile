ARG FUNCTION_IMG
ARG FROM=node:19

FROM scratch as server
ARG FILES=node-static-builder
COPY ${FILES} /app

FROM ${FUNCTION_IMG} as function
FROM ${FROM} as build

RUN mkdir -p /app

COPY --from=server /app /app
RUN rm -rf /app/function
COPY --from=function /app /app/function

WORKDIR /app/function
RUN npm ci
RUN npm run build

FROM nginx:stable

COPY --from=server /app/nginx.conf /etc/nginx/nginx.conf
RUN mkdir -p /app
WORKDIR /app
COPY --from=build /app/function/build /app/build

CMD ["nginx"]