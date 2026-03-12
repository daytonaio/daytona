FROM node:22-alpine AS build

RUN npm install -g corepack && corepack enable

COPY --from=golang:1.23.5-alpine /usr/local/go/ /usr/local/go/

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
COPY apps/runner/ apps/runner/
COPY apps/daemon/ apps/daemon/
COPY libs/common-go/ libs/common-go/
COPY libs/api-client-go/ libs/api-client-go/
COPY libs/computer-use/ libs/computer-use/
RUN head -1 go.work > go.work.tmp && printf '\nuse (\n\t./apps/runner\n\t./apps/daemon\n\t./libs/common-go\n\t./libs/api-client-go\n)\n' >> go.work.tmp && mv go.work.tmp go.work

# Pre-built computer-use binary and build script
COPY dist/libs/computer-use-amd64 dist/libs/computer-use-amd64

ARG VERSION=0.0.1
ENV VERSION=${VERSION}
ENV NX_DAEMON=false

ENV GONOSUMDB=github.com/daytonaio/daytona

RUN SKIP_COMPUTER_USE_BUILD=true yarn nx build runner --configuration=production --nxBail=true

FROM docker:28.2.2-dind-alpine3.22 AS runner

RUN apk add --no-cache curl rsync

WORKDIR /usr/local/bin

COPY --from=build /daytona/dist/apps/runner daytona-runner

RUN chmod +x daytona-runner

RUN mkdir -p /etc/docker && echo '{"insecure-registries": ["registry:6000"]}' > /etc/docker/daemon.json

HEALTHCHECK CMD [ "curl", "-f", "http://localhost:3003/" ]

ENTRYPOINT ["sh", "-c", "/usr/local/bin/dockerd-entrypoint.sh & daytona-runner"]
