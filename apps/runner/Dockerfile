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
COPY apps/runner/go.mod apps/runner/go.sum apps/runner/
COPY apps/daemon/go.mod apps/daemon/go.sum apps/daemon/
COPY libs/common-go/go.mod libs/common-go/go.sum libs/common-go/
COPY libs/api-client-go/go.mod libs/api-client-go/go.sum libs/api-client-go/
RUN head -1 go.work > go.work.tmp && printf '\nuse (\n\t./apps/runner\n\t./apps/daemon\n\t./libs/common-go\n\t./libs/api-client-go\n)\n' >> go.work.tmp && mv go.work.tmp go.work

ENV NX_DAEMON=false
ENV GONOSUMDB=github.com/daytonaio/daytona

RUN go -C apps/runner mod download \
  && go -C apps/daemon mod download \
  && go -C libs/common-go mod download \
  && go -C libs/api-client-go mod download

# Go source
COPY apps/runner/ apps/runner/
COPY apps/daemon/ apps/daemon/
COPY libs/common-go/ libs/common-go/
COPY libs/api-client-go/ libs/api-client-go/
COPY libs/computer-use/ libs/computer-use/

# Pre-built computer-use binary and build script
COPY dist/libs/computer-use-amd64 dist/libs/computer-use-amd64

# Pin google.golang.org/genproto to a post-split version. The trimmed workspace
# lacks snapshot-manager -> distribution/v3 which normally bumps genproto past the
# split. Without this, in-toto-golang pulls in a pre-split version that conflicts
# with the split genproto/googleapis/{api,rpc} modules.
RUN GOWORK=off go -C apps/runner mod edit -require google.golang.org/genproto@v0.0.0-20240903143218-8af14fe29dc1

ARG VERSION=0.0.1
RUN --mount=type=cache,target=/root/.cache/go-build \
  SKIP_COMPUTER_USE_BUILD=true VERSION=$VERSION yarn nx build runner --configuration=production --nxBail=true

FROM docker:28.2.2-dind-alpine3.22 AS runner

RUN apk add --no-cache curl rsync

WORKDIR /usr/local/bin

COPY --from=build /daytona/dist/apps/runner daytona-runner

RUN chmod +x daytona-runner

RUN mkdir -p /etc/docker && echo '{"insecure-registries": ["registry:6000"]}' > /etc/docker/daemon.json

HEALTHCHECK CMD [ "curl", "-f", "http://localhost:3003/" ]

ENTRYPOINT ["sh", "-c", "/usr/local/bin/dockerd-entrypoint.sh & daytona-runner"]
