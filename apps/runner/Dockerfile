FROM node:22-alpine AS build

RUN npm install -g corepack && corepack enable

COPY --from=golang:1.23.5-alpine /usr/local/go/ /usr/local/go/

ENV PATH="/usr/local/go/bin:${PATH}"

WORKDIR /daytona

COPY . .

ARG VERSION=0.0.1
ENV VERSION=${VERSION}

RUN yarn

RUN SKIP_COMPUTER_USE_BUILD=true yarn nx build runner --configuration=production --nxBail=true

FROM docker:28.2.2-dind-alpine3.22 AS runner

WORKDIR /usr/local/bin

COPY --from=build /daytona/dist/apps/runner daytona-runner

RUN chmod +x daytona-runner

RUN mkdir -p /etc/docker && echo '{"insecure-registries": ["registry:6000"]}' > /etc/docker/daemon.json

ENV ENVIRONMENT=development

ENV API_PORT=3003
ENV API_TOKEN=secret_api_token

ENV DAYTONA_BINARY_PATH=/workspaces/daytona/dist/apps/daemon-amd64
ENV LOG_FILE_PATH=/home/daytona/runner/runner.log

ENV RESOURCE_LIMITS_DISABLED=true

ENV AWS_ENDPOINT_URL=http://minio:9000
ENV AWS_REGION=us-east-1
ENV AWS_ACCESS_KEY_ID=minioadmin
ENV AWS_SECRET_ACCESS_KEY=minioadmin
ENV AWS_DEFAULT_BUCKET=daytona

ENTRYPOINT ["sh", "-c", "dockerd & daytona-runner"]
