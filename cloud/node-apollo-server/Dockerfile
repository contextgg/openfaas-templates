FROM node:lts-alpine

RUN mkdir -p /app
WORKDIR /app

ENV NODE_ENV=production

COPY . .
RUN npm install

CMD ["npm", "run", "start"]