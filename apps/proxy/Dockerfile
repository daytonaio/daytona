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

RUN yarn nx build proxy --configuration=production --nxBail=true

FROM alpine:3.18 AS proxy

RUN apk add --no-cache curl

WORKDIR /usr/local/bin

COPY --from=build /daytona/dist/apps/proxy daytona-proxy

RUN chmod +x daytona-proxy

HEALTHCHECK CMD [ "curl", "-f", "http://localhost:4000/health" ]

ENTRYPOINT ["daytona-proxy"]
