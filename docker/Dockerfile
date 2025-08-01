FROM docker:28.2.2-dind-alpine3.22 AS daytona

# Install nodejs
RUN apk --update add --no-cache nodejs=22.16.0-r2 npm bash
RUN npm install -g corepack && corepack enable

COPY --from=golang:1.23.5-alpine /usr/local/go/ /usr/local/go/

ENV PATH="/usr/local/go/bin:${PATH}"

WORKDIR /daytona

COPY . .

# Docs build arguments
ENV PUBLIC_WEB_URL=https://daytona.io
ARG PUBLIC_ALGOLIA_APP_ID
ENV PUBLIC_ALGOLIA_APP_ID=${PUBLIC_ALGOLIA_APP_ID}
ARG PUBLIC_ALGOLIA_API_KEY
ENV PUBLIC_ALGOLIA_API_KEY=${PUBLIC_ALGOLIA_API_KEY}

ARG VERSION=0.0.1
ENV VERSION=${VERSION}

RUN yarn && \
  SKIP_COMPUTER_USE_BUILD=true yarn build:production-skip-computer-use --nxBail=true

RUN mkdir -p /etc/docker && echo '{"iptables": false, "ip-forward": false, "ip-masq": false, "userland-proxy": false,"insecure-registries": ["localhost:6000"]}' > /etc/docker/daemon.json

ENV PORT=3000

ENV DB_HOST=db
ENV DB_PORT=5432
ENV DB_USERNAME=user
ENV DB_PASSWORD=pass
ENV DB_DATABASE=daytona

ENV REDIS_HOST=redis
ENV REDIS_PORT=6379

ENV OIDC_CLIENT_ID=daytona
ENV OIDC_ISSUER_BASE_URL=http://localhost:5556/dex
ENV OIDC_AUDIENCE=daytona

ENV OIDC_MANAGEMENT_API_ENABLED=
ENV OIDC_MANAGEMENT_API_CLIENT_ID=
ENV OIDC_MANAGEMENT_API_CLIENT_SECRET=
ENV OIDC_MANAGEMENT_API_AUDIENCE=

ENV DEFAULT_SNAPSHOT=ubuntu:22.04
ENV DASHBOARD_URL=http://localhost:3000/dashboard

ENV POSTHOG_API_KEY=phc_bYtEsdMDrNLydXPD4tufkBrHKgfO2zbycM30LOowYNv
ENV POSTHOG_HOST=https://d18ag4dodbta3l.cloudfront.net
ENV POSTHOG_ENVIRONMENT=local

ENV TRANSIENT_REGISTRY_URL=http://localhost:6000
ENV TRANSIENT_REGISTRY_ADMIN=admin
ENV TRANSIENT_REGISTRY_PASSWORD=password
ENV TRANSIENT_REGISTRY_PROJECT_ID=daytona

ENV INTERNAL_REGISTRY_URL=http://localhost:6000
ENV INTERNAL_REGISTRY_ADMIN=admin
ENV INTERNAL_REGISTRY_PASSWORD=password
ENV INTERNAL_REGISTRY_PROJECT_ID=daytona

ENV SMTP_HOST=maildev
ENV SMTP_PORT=1025
ENV SMTP_USER=
ENV SMTP_PASSWORD=
ENV SMTP_SECURE=
ENV SMTP_EMAIL_FROM="Daytona Team <no-reply@daytona.io>"

ENV S3_ENDPOINT=http://minio:9000
ENV S3_STS_ENDPOINT=http://minio:9000/minio/v1/assume-role
ENV S3_REGION=us-east-1
ENV S3_ACCESS_KEY=minioadmin
ENV S3_SECRET_KEY=minioadmin
ENV S3_DEFAULT_BUCKET=daytona
ENV S3_ACCOUNT_ID=/
ENV S3_ROLE_NAME=/

ENV ENVIRONMENT=dev

ENV MAX_AUTO_ARCHIVE_INTERVAL=43200

ENV OTEL_ENABLED=true
ENV OTEL_COLLECTOR_URL=http://jaeger:4318/v1/traces

ENV MAINTENANCE_MODE=false

ENV PROXY_DOMAIN=proxy.localhost:4000
ENV PROXY_PROTOCOL=http
ENV PROXY_API_KEY=super_secret_key
ENV PROXY_TEMPLATE_URL=http://{{PORT}}-{{sandboxId}}.proxy.localhost:4000

ENV DEFAULT_RUNNER_DOMAIN=localhost:3003
ENV DEFAULT_RUNNER_API_URL=http://localhost:3003
ENV DEFAULT_RUNNER_PROXY_URL=http://localhost:3003
ENV DEFAULT_RUNNER_API_KEY=secret_api_token
ENV DEFAULT_RUNNER_CPU=4
ENV DEFAULT_RUNNER_MEMORY=8
ENV DEFAULT_RUNNER_DISK=50
ENV DEFAULT_RUNNER_GPU=0
ENV DEFAULT_RUNNER_GPU_TYPE=none
ENV DEFAULT_RUNNER_CAPACITY=100
ENV DEFAULT_RUNNER_REGION=us
ENV DEFAULT_RUNNER_CLASS=small

ENTRYPOINT ["sh", "-c", "dockerd-entrypoint.sh & node dist/apps/api/main.js"]

FROM alpine:3.18 AS proxy

WORKDIR /usr/local/bin

COPY --from=daytona /daytona/dist/apps/proxy daytona-proxy

RUN chmod +x daytona-proxy

ENV DAYTONA_API_URL=http://localhost:3000/api

ENV PROXY_PORT=4000
ENV PROXY_DOMAIN=proxy.localhost:4000
ENV PROXY_API_KEY=super_secret_key
ENV PROXY_PROTOCOL=http

ENV OIDC_CLIENT_ID=daytona
ENV OIDC_CLIENT_SECRET=
ENV OIDC_DOMAIN=http://localhost:5556/dex
ENV OIDC_AUDIENCE=daytona

ENV REDIS_HOST=redis
ENV REDIS_PORT=6379

ENTRYPOINT ["daytona-proxy"]

FROM docker:28.2.2-dind-alpine3.22 AS runner

WORKDIR /usr/local/bin

COPY --from=daytona /daytona/dist/apps/runner daytona-runner

RUN chmod +x daytona-runner

RUN mkdir -p /etc/docker && echo '{"iptables": false, "ip-forward": false, "ip-masq": false, "userland-proxy": false,"insecure-registries": ["localhost:6000"]}' > /etc/docker/daemon.json

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

FROM node:22-alpine AS docs

WORKDIR /daytona

COPY --from=daytona /daytona/node_modules node_modules
COPY --from=daytona /daytona/dist/apps/docs dist/apps/docs
COPY --from=daytona /daytona/apps/docs/server dist/apps/docs/server

WORKDIR /daytona/dist/apps/docs

ENTRYPOINT ["sh", "-c", "node server/index.mjs"]