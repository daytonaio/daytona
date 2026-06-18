FROM node:20-alpine AS builder

ENV CI=true

# Build toolchain (python3/py3-setuptools/make/g++) is needed when a native
# module's prebuilt binary download fails and node-gyp falls back to compiling
# from source — it stays out of the runtime image. py3-setuptools provides the
# distutils module node-gyp needs on Python 3.12+.
RUN apk add --no-cache python3 py3-setuptools make g++

RUN npm install -g corepack && corepack enable

WORKDIR /daytona

# Yarn caching layer
COPY package.json yarn.lock .yarnrc.yml ./
RUN yarn install --immutable

# Nx + TS config
COPY nx.json tsconfig.base.json .nxignore ./

# App source
COPY apps/dashboard/ apps/dashboard/

# Lib dependencies
COPY libs/api-client/ libs/api-client/
COPY libs/analytics-api-client/ libs/analytics-api-client/
COPY libs/toolbox-api-client/ libs/toolbox-api-client/
COPY libs/sdk-typescript/ libs/sdk-typescript/
COPY libs/billing-api-client/ libs/billing-api-client/

ENV NX_DAEMON=false

RUN VITE_BASE_PATH=/dashboard/ VITE_API_URL=%DAYTONA_BASE_API_URL%/api yarn nx build dashboard --configuration=production --nxBail=true

FROM nginx:alpine as dashboard

COPY --from=builder /daytona/dist/apps/dashboard /usr/share/nginx/html/dashboard
COPY --from=builder /daytona/apps/dashboard/docker/nginx/default.conf /etc/nginx/conf.d/default.conf
COPY --from=builder /daytona/apps/dashboard/docker/entrypoint.sh /entrypoint.sh

ENV DAYTONA_BASE_API_URL="http://api:3000"

EXPOSE 80

HEALTHCHECK CMD [ "wget", "-q", "--spider", "http://localhost/" ]

ENTRYPOINT ["/entrypoint.sh"]
