FROM node:22-alpine AS build

RUN npm install -g corepack && corepack enable

COPY --from=golang:1.23.5-alpine /usr/local/go/ /usr/local/go/

ENV PATH="/usr/local/go/bin:${PATH}"

RUN apk add --no-cache git

WORKDIR /daytona

COPY . .

ARG VERSION=0.0.1
ENV VERSION=${VERSION}

ENV GONOSUMDB=github.com/daytonaio/daytona

RUN yarn

RUN yarn nx build ssh-gateway --configuration=production --nxBail=true

FROM alpine:3.18 AS ssh-gateway

WORKDIR /usr/local/bin

COPY --from=build /daytona/dist/apps/ssh-gateway daytona-ssh-gateway

RUN chmod +x daytona-ssh-gateway

EXPOSE 2222

ENTRYPOINT ["daytona-ssh-gateway"]
