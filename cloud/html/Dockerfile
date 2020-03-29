FROM nginx:stable

RUN mkdir -p /app
WORKDIR /app

CMD ["nginx"]

COPY nginx.conf /etc/nginx/nginx.conf
COPY function/ /app/build