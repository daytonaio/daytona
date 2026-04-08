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

# Go dependency layer (cached unless go.mod/go.sum change)
COPY go.work go.work.sum ./
COPY apps/snapshot-manager/go.mod apps/snapshot-manager/go.sum apps/snapshot-manager/
RUN head -1 go.work > go.work.tmp && printf '\nuse (\n\t./apps/snapshot-manager\n)\n' >> go.work.tmp && mv go.work.tmp go.work

ENV NX_DAEMON=false
ENV GONOSUMDB=github.com/daytonaio/daytona

RUN cd apps/snapshot-manager && go mod download

# Go source
COPY apps/snapshot-manager/ apps/snapshot-manager/

ARG VERSION=0.0.1
RUN --mount=type=cache,target=/root/.cache/go-build \
  VERSION=$VERSION yarn nx build snapshot-manager --configuration=production --nxBail=true

FROM alpine:3.22 AS snapshot-manager

RUN apk add --no-cache curl

WORKDIR /usr/local/bin

COPY --from=build /daytona/dist/apps/snapshot-manager daytona-snapshot-manager

RUN chmod +x daytona-snapshot-manager

ENTRYPOINT ["daytona-snapshot-manager"]
