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
COPY apps/otel-collector/ apps/otel-collector/
COPY libs/api-client-go/ libs/api-client-go/
COPY libs/common-go/ libs/common-go/
RUN head -1 go.work > go.work.tmp && printf '\nuse (\n\t./apps/otel-collector/exporter\n\t./libs/api-client-go\n\t./libs/common-go\n)\n' >> go.work.tmp && mv go.work.tmp go.work
RUN mkdir -p dist/apps/otel-collector

ARG VERSION=0.0.1
ENV VERSION=${VERSION}
ENV NX_DAEMON=false

ENV GONOSUMDB=github.com/daytonaio/daytona

RUN yarn nx build otel-collector --configuration=production --nxBail=true

FROM alpine:3.18 AS otel-collector

RUN apk add --no-cache curl

WORKDIR /usr/local/bin

COPY --from=build /daytona/dist/apps/otel-collector/daytona-otel-collector daytona-otel-collector
COPY --from=build /daytona/apps/otel-collector/config.yaml /otelcol/collector-config.yaml

RUN chmod +x daytona-otel-collector

HEALTHCHECK CMD [ "curl", "-f", "http://localhost:13133/health/status" ]

EXPOSE 4317
EXPOSE 4318
EXPOSE 13133
EXPOSE 8888

ENTRYPOINT ["daytona-otel-collector"]

CMD ["--config", "/otelcol/collector-config.yaml"]
