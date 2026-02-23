FROM node:22-alpine AS build

RUN npm install -g corepack && corepack enable

COPY --from=golang:1.23.5-alpine /usr/local/go/ /usr/local/go/

ENV PATH="/usr/local/go/bin:${PATH}"

WORKDIR /daytona

COPY . .

ARG VERSION=0.0.1
ENV VERSION=${VERSION}

ENV GONOSUMDB=github.com/daytonaio/daytona

RUN yarn

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