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
COPY apps/proxy/go.mod apps/proxy/go.sum apps/proxy/
COPY libs/common-go/go.mod libs/common-go/go.sum libs/common-go/
COPY libs/api-client-go/go.mod libs/api-client-go/go.sum libs/api-client-go/
RUN head -1 go.work > go.work.tmp && printf '\nuse (\n\t./apps/proxy\n\t./libs/common-go\n\t./libs/api-client-go\n)\n' >> go.work.tmp && mv go.work.tmp go.work

ENV NX_DAEMON=false
ENV GONOSUMDB=github.com/daytonaio/daytona

RUN go -C apps/proxy mod download && go -C libs/common-go mod download && go -C libs/api-client-go mod download

# Go source
COPY apps/proxy/ apps/proxy/
COPY libs/common-go/ libs/common-go/
COPY libs/api-client-go/ libs/api-client-go/

ARG VERSION=0.0.1
RUN --mount=type=cache,target=/root/.cache/go-build \
  VERSION=$VERSION yarn nx build proxy --configuration=production --nxBail=true

FROM alpine:3.22 AS proxy

RUN apk add --no-cache curl

WORKDIR /usr/local/bin

COPY --from=build /daytona/dist/apps/proxy daytona-proxy

RUN chmod +x daytona-proxy

HEALTHCHECK CMD [ "curl", "-f", "http://localhost:4000/health" ]

ENTRYPOINT ["daytona-proxy"]
