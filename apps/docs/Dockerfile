FROM node:20.10.0 as deps

WORKDIR /usr/src/app

COPY package.json yarn.lock ./

RUN yarn --prod

FROM node:20.10.0 as build

ARG PUBLIC_WEB_URL

ENV PUBLIC_WEB_URL=${PUBLIC_WEB_URL}

WORKDIR /usr/src/app

COPY . .

RUN yarn
RUN yarn build

FROM node:20.10.0-slim

WORKDIR /usr/src/app

COPY server ./server
COPY --link --from=build /usr/src/app/dist ./dist
COPY --link --from=deps /usr/src/app/node_modules ./node_modules

CMD [ "node", "server/index.mjs" ]
