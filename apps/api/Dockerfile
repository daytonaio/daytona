FROM node:24-slim AS daytona

ENV CI=true

# Install dependencies (apt instead of apk)
RUN apt-get update && apt-get install -y --no-install-recommends bash curl python3 make g++ && \
  rm -rf /var/lib/apt/lists/*
RUN npm install -g corepack && corepack enable

WORKDIR /daytona

# Yarn caching layer
COPY package.json yarn.lock .yarnrc.yml ./
RUN yarn install --immutable

# Nx + TS config
COPY nx.json tsconfig.base.json ./

# App source
COPY apps/api/ apps/api/
COPY apps/dashboard/ apps/dashboard/

# Lib dependencies
COPY libs/runner-api-client/ libs/runner-api-client/
COPY libs/api-client/ libs/api-client/
COPY libs/analytics-api-client/ libs/analytics-api-client/
COPY libs/toolbox-api-client/ libs/toolbox-api-client/
COPY libs/sdk-typescript/ libs/sdk-typescript/
COPY libs/runner-proto/ libs/runner-proto/

ENV NX_DAEMON=false

RUN yarn nx build api --configuration=production --nxBail=true

RUN VITE_API_URL=%DAYTONA_BASE_API_URL%/api yarn nx build dashboard --configuration=production --nxBail=true --output-style=stream

ARG VERSION=0.0.1
ENV VERSION=${VERSION}

HEALTHCHECK CMD [ "curl", "-f", "http://localhost:3000/api/config" ]

ENTRYPOINT ["node", "dist/apps/api/main.js"]
