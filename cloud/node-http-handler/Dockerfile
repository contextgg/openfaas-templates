FROM node:lts-alpine

RUN mkdir -p /app
WORKDIR /app

ENV NODE_ENV=production

COPY function/package.json ./
COPY function/package*.json ./
RUN npm install

COPY function/ ./
RUN npm run build

CMD ["npm", "run", "start"]