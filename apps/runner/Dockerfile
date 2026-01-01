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
 
# Install dependencies
RUN apk add --no-cache \
    curl \
    rsync \
    fuse3 \
    ca-certificates \
    tar

# Install AWS Mountpoint for S3 (mount-s3) from official AWS releases
RUN set -eux; \
    ARCH="$(uname -m)"; \
    case "$ARCH" in \
      x86_64) ARCH=x86_64 ;; \
      aarch64) ARCH=aarch64 ;; \
      *) echo "Unsupported architecture: $ARCH" && exit 1 ;; \
    esac; \
    curl -fL "https://s3.amazonaws.com/mountpoint-s3-release/latest/mountpoint-s3-linux-${ARCH}.tar.gz" \
      -o /tmp/mountpoint-s3.tar.gz; \
    tar -xzf /tmp/mountpoint-s3.tar.gz -C /tmp; \
    mv /tmp/mount-s3 /usr/local/bin/mount-s3; \
    chmod +x /usr/local/bin/mount-s3; \
    rm -rf /tmp/mountpoint-s3.tar.gz


WORKDIR /usr/local/bin

COPY --from=build /daytona/dist/apps/runner daytona-runner

RUN chmod +x daytona-runner

RUN mkdir -p /etc/docker && echo '{"insecure-registries": ["registry:6000"]}' > /etc/docker/daemon.json

ENV ENVIRONMENT=development

ENV API_PORT=3003
ENV API_TOKEN=secret_api_token

ENV LOG_FILE_PATH=/home/daytona/runner/runner.log

ENV RESOURCE_LIMITS_DISABLED=true

ENV AWS_ENDPOINT_URL=http://minio:9000
ENV AWS_REGION=us-east-1
ENV AWS_ACCESS_KEY_ID=minioadmin
ENV AWS_SECRET_ACCESS_KEY=minioadmin
ENV AWS_DEFAULT_BUCKET=daytona

ENV SERVER_URL=http://api:3000/api

ENV SSH_GATEWAY_ENABLE=true
ENV SSH_PUBLIC_KEY="c3NoLXJzYSBBQUFBQjNOemFDMXljMkVBQUFBREFRQUJBQUFCZ1FEZUtocFdOYVhVNkpSMnJXekc1NlZJeksrcmsrdForcCtVUmwyeWF5cWRnZGg2bnVOL1NtVnJsYmV4MGlrRVZZemdRWUIwWno5M2dqOVhxSkd4WjJiMHVoSDFzNkp2bnhKZVdIK05rbjBNMzRiRjZkeVloZ3o4M0JaTitWcElJdnVDT1pxYWE4VnlRWjhPdjVLMU00enlmQnFCMld3ejcwZEVWUnFaMnVZckd5U0RqUUIrU3hRQ3phY0djb2JRQTFUd2QxdEJvSk5BaVpPWXFSU3d6TmNoTk5hZHlxSHRGdkZRQ0hhNjZWaVBmQ0FVUGt2cEN6ZGh3TmJGdVhGVWkxbnZOWkx4M056RU41M05LWUhZcWU3Y3dXZkI4QjBoRDhDd3hmSDd6T0o1b1RCMnBMVVlsUWpZRDNsN3k4SkFTbDZpYkpuR1A1SWczVXpZWlRIdUdkNzVXWktnNTRJTVROQlFPMWpJdE9HL0orUm1XbzB0YTdpRkF6ZExyL1BmbDdXTVMwbEZhSm9scTdUVmJLUG1JODFTOU04VEgrbXhLbGZ5b2NqZzhwSUdlQUllcEo3dXl1Vk1xT2wxeVVuSFNycDRlVEVlR0k3NlpCOFRoUzhkYnBLUTIvb2RYMHkwc3FSZDI0Y2lGdnM0dnZVaW80NFdYNlNWRG54dXpWRHc1Rzg9IGRheXRvbmFAMDQwNjZiZDIwY2Vi"

HEALTHCHECK CMD [ "curl", "-f", "http://localhost:3003/" ]

ENTRYPOINT ["sh", "-c", "/usr/local/bin/dockerd-entrypoint.sh & daytona-runner"]
