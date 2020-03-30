FROM node:lts-alpine

RUN mkdir -p /app
WORKDIR /app

COPY package.json .
COPY package-lock.json .
RUN npm install

COPY . .
RUN npm run bootstrap

ENV NODE_ENV=production
RUN npm prune

CMD ["npm", "run", "start"]