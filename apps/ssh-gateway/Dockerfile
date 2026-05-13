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
COPY apps/ssh-gateway/go.mod apps/ssh-gateway/go.sum apps/ssh-gateway/
COPY libs/api-client-go/go.mod libs/api-client-go/go.sum libs/api-client-go/
RUN head -1 go.work > go.work.tmp && printf '\nuse (\n\t./apps/ssh-gateway\n\t./libs/api-client-go\n)\n' >> go.work.tmp && mv go.work.tmp go.work

ENV NX_DAEMON=false
ENV GONOSUMDB=github.com/daytonaio/daytona

RUN go -C apps/ssh-gateway mod download && go -C libs/api-client-go mod download

# Go source
COPY apps/ssh-gateway/ apps/ssh-gateway/
COPY libs/api-client-go/ libs/api-client-go/

RUN yarn nx build ssh-gateway --configuration=production --nxBail=true

FROM alpine:3.22 AS ssh-gateway

WORKDIR /usr/local/bin

COPY --from=build /daytona/dist/apps/ssh-gateway daytona-ssh-gateway

RUN chmod +x daytona-ssh-gateway

EXPOSE 2222

ENTRYPOINT ["daytona-ssh-gateway"]
