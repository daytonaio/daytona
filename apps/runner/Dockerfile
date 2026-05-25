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
COPY nx.json .nxignore ./

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

# Build mount-s3 (Mountpoint for Amazon S3) from source against Alpine musl.
# AWS only publishes glibc-linked mount-s3 artifacts (the .deb/tar.gz on the
# Mountpoint releases page), which cannot run on Alpine. Pinned to 1.20.0 to
# match the version already pinned in .devcontainer/Dockerfile.
FROM alpine:3.22 AS mount-s3-builder
ARG MOUNTPOINT_S3_VERSION=1.20.0
RUN apk add --no-cache \
      git curl gcc g++ make cmake pkgconfig perl \
      zlib zlib-dev openssl-dev fuse fuse-dev \
      clang clang-dev rust cargo musl-dev linux-headers
RUN git clone --recurse-submodules --shallow-submodules --depth 1 \
      --branch mountpoint-s3-${MOUNTPOINT_S3_VERSION} \
      https://github.com/awslabs/mountpoint-s3.git /src/mountpoint-s3
RUN --mount=type=cache,target=/root/.cargo/registry \
    --mount=type=cache,target=/src/mountpoint-s3/target,sharing=locked \
    cd /src/mountpoint-s3 \
 && cargo build --release --locked --bin mount-s3 \
 && cp /src/mountpoint-s3/target/release/mount-s3 /usr/local/bin/mount-s3 \
 && strip /usr/local/bin/mount-s3 \
 && /usr/local/bin/mount-s3 --version

FROM docker:28.5.2-dind-alpine3.22 AS runner

RUN apk update && apk upgrade --no-cache openssl pcre2 \
  && apk add --no-cache curl rsync fuse libgcc ca-certificates

WORKDIR /usr/local/bin

COPY --from=build /daytona/dist/apps/runner daytona-runner
COPY --from=mount-s3-builder --chmod=0755 /usr/local/bin/mount-s3 /usr/local/bin/mount-s3

RUN chmod +x daytona-runner

# Verify mount-s3 loads against the final-stage runtime libraries.
RUN /usr/local/bin/mount-s3 --version

RUN mkdir -p /etc/docker && echo '{"insecure-registries": ["registry:6000"]}' > /etc/docker/daemon.json

HEALTHCHECK CMD [ "curl", "-f", "http://localhost:3003/" ]

ENTRYPOINT ["sh", "-c", "/usr/local/bin/dockerd-entrypoint.sh & daytona-runner"]
