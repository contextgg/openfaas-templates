FROM node:lts-alpine

RUN addgroup -S app && \
  adduser -S -G app app && \
  mkdir -p /app && \
  chmod 775 /app && \
  chown app:app /app
WORKDIR /app
USER app

COPY . .
RUN npm install

ENV NODE_ENV=production
RUN npm prune

CMD ["npm", "run", "start"]