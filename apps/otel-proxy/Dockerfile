FROM node:22-alpine AS build

RUN npm install -g corepack && corepack enable

COPY --from=golang:1.23.5-alpine /usr/local/go/ /usr/local/go/

ENV PATH="/usr/local/go/bin:${PATH}"

WORKDIR /daytona

COPY . .

ARG VERSION=0.0.1
ENV VERSION=${VERSION}

RUN yarn

RUN yarn nx build otel-proxy --configuration=production --nxBail=true

FROM alpine:3.18 AS otel-proxy

WORKDIR /usr/local/bin

COPY --from=build /daytona/dist/apps/otel-proxy daytona-otel-proxy

RUN chmod +x daytona-otel-proxy

ENV DAYTONA_API_URL=http://api:3000/api

ENV PROXY_PORT=4000
ENV PROXY_DOMAIN=proxy.localhost:4000
ENV PROXY_API_KEY=super_secret_key
ENV PROXY_PROTOCOL=http

ENV OIDC_CLIENT_ID=daytona
ENV OIDC_CLIENT_SECRET=
ENV OIDC_DOMAIN=http://dex:5556/dex
ENV OIDC_PUBLIC_DOMAIN=http://localhost:5556/dex
ENV OIDC_AUDIENCE=daytona

ENV REDIS_HOST=redis
ENV REDIS_PORT=6379

ENTRYPOINT ["daytona-otel-proxy"]
