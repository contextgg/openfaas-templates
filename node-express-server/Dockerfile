FROM node:lts-alpine

RUN mkdir -p /app
WORKDIR /app

COPY package.json .
COPY package-lock.json .
COPY lerna.json .
RUN npm install

COPY function/package.json function/
COPY function/package*.json function/

RUN npm run bootstrap

COPY . .

RUN npm --prefix /app/function run build --if-present

ENV NODE_ENV=production
# RUN npm prune

EXPOSE 8080/TCP
EXPOSE 8081/TCP
EXPOSE 8082/TCP

CMD ["npm", "run", "start"]