FROM node:22-alpine AS build

ENV CI=true

RUN npm install -g corepack && corepack enable

COPY --from=golang:1.25.4-alpine /usr/local/go/ /usr/local/go/

ENV PATH="/usr/local/go/bin:${PATH}"

RUN apk add --no-cache git

WORKDIR /daytona

# Yarn caching layer
COPY package.json yarn.lock .yarnrc.yml ./
RUN yarn install --immutable

# Nx config
COPY nx.json ./

# Go source
COPY go.work go.work.sum ./
COPY apps/proxy/ apps/proxy/
COPY libs/common-go/ libs/common-go/
COPY libs/api-client-go/ libs/api-client-go/
RUN head -1 go.work > go.work.tmp && printf '\nuse (\n\t./apps/proxy\n\t./libs/common-go\n\t./libs/api-client-go\n)\n' >> go.work.tmp && mv go.work.tmp go.work

ARG VERSION=0.0.1
ENV VERSION=${VERSION}
ENV NX_DAEMON=false

ENV GONOSUMDB=github.com/daytonaio/daytona

RUN yarn nx build proxy --configuration=production --nxBail=true

FROM alpine:3.18 AS proxy

RUN apk add --no-cache curl

WORKDIR /usr/local/bin

COPY --from=build /daytona/dist/apps/proxy daytona-proxy

RUN chmod +x daytona-proxy

HEALTHCHECK CMD [ "curl", "-f", "http://localhost:4000/health" ]

ENTRYPOINT ["daytona-proxy"]
