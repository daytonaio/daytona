FROM golang:1.25.11-alpine AS go-toolchain

FROM node:22-alpine AS build

ENV CI=true

# Build toolchain (python3/py3-setuptools/make/g++) is needed when a native
# module's prebuilt binary download fails and node-gyp falls back to compiling
# from source — it stays out of the runtime image. py3-setuptools provides the
# distutils module node-gyp needs on Python 3.12+.
RUN apk add --no-cache python3 py3-setuptools make g++

RUN npm install -g corepack && corepack enable

COPY --from=go-toolchain /usr/local/go/ /usr/local/go/

ENV PATH="/usr/local/go/bin:${PATH}"

RUN apk add --no-cache git

WORKDIR /daytona

# Yarn caching layer
COPY package.json yarn.lock .yarnrc.yml ./
RUN yarn install --immutable

# Nx config
COPY nx.json .nxignore ./

# Go dependency layer (cached unless go.mod/go.sum change)
COPY go.work go.work.sum ./
COPY apps/ssh-gateway/go.mod apps/ssh-gateway/go.sum apps/ssh-gateway/
COPY libs/api-client-go/go.mod libs/api-client-go/go.sum libs/api-client-go/
RUN head -1 go.work > go.work.tmp && printf '\nuse (\n\t./apps/ssh-gateway\n\t./libs/api-client-go\n)\n' >> go.work.tmp && mv go.work.tmp go.work

ENV NX_DAEMON=false
ENV GONOSUMDB=github.com/daytonaio/daytona

RUN go -C apps/ssh-gateway mod download && go -C libs/api-client-go mod download

# Go source
COPY apps/ssh-gateway/ apps/ssh-gateway/
COPY libs/api-client-go/ libs/api-client-go/

RUN yarn nx build ssh-gateway --configuration=production --nxBail=true

FROM alpine:3.23 AS ssh-gateway

RUN apk update && apk upgrade --no-cache openssl pcre2

WORKDIR /usr/local/bin

COPY --from=build /daytona/dist/apps/ssh-gateway daytona-ssh-gateway

RUN chmod +x daytona-ssh-gateway

EXPOSE 2222

ENTRYPOINT ["daytona-ssh-gateway"]
