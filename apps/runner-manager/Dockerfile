FROM node:22-alpine AS build

RUN npm install -g corepack && corepack enable

COPY --from=golang:1.23.5-alpine /usr/local/go/ /usr/local/go/

ENV PATH="/usr/local/go/bin:${PATH}"

WORKDIR /daytona

COPY . .

ARG VERSION=0.0.1
ENV VERSION=${VERSION}

RUN yarn

RUN yarn nx build runner-manager --configuration=production --nxBail=true

FROM alpine:3.22 AS runner-manager

RUN apk add --no-cache curl ca-certificates

WORKDIR /usr/local/bin

COPY --from=build /daytona/dist/apps/runner-manager daytona-runner-manager

RUN chmod +x daytona-runner-manager

ENV ENVIRONMENT=development
ENV API_PORT=3010
ENV LOG_LEVEL=info

HEALTHCHECK CMD [ "curl", "-f", "http://localhost:3010/health" ]

ENTRYPOINT ["daytona-runner-manager"]





