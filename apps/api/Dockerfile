# ── builder ──────────────────────────────────────────────────────────────
FROM node:24-slim AS builder

ENV CI=true

# Build toolchain (python3/make/g++) is only needed to compile the bundle and
# any native devDependencies — it stays out of the runtime image.
RUN apt-get update && apt-get install -y --no-install-recommends \
      bash curl python3 make g++ && \
    rm -rf /var/lib/apt/lists/*
RUN npm install -g corepack && corepack enable

WORKDIR /daytona

# Yarn caching layer
COPY package.json yarn.lock .yarnrc.yml ./
RUN yarn install --immutable

# Nx + TS config
COPY nx.json tsconfig.base.json .nxignore ./

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
COPY libs/billing-api-client/ libs/billing-api-client/

ENV NX_DAEMON=false

RUN yarn nx build api --configuration=production --nxBail=true

RUN VITE_API_URL=%DAYTONA_BASE_API_URL%/api yarn nx build dashboard --configuration=production --nxBail=true --output-style=stream

# ── runtime ──────────────────────────────────────────────────────────────
FROM node:24-slim AS daytona

ENV CI=true

RUN apt-get update && apt-get install -y --no-install-recommends curl && \
    rm -rf /var/lib/apt/lists/*
RUN npm install -g corepack && corepack enable

WORKDIR /daytona

# Nx (generatePackageJson) emits a pruned manifest + pruned lockfile carrying
# only the api bundle's runtime closure. apps/api/package.json augments it with
# runtime deps Nx's static scan misses: @openfeature/core (a non-optional peer
# of @openfeature/server-sdk) and pg (typeorm's Postgres driver, loaded via a
# dynamic require). It also adds the typeorm-migration toolchain (ts-node,
# tsconfig-paths, typescript, reflect-metadata, dotenv) plus the migration:run:*
# scripts, so the deploy migration init/pre/post-deploy jobs run from this image.
# corepack hydrates yarn from the manifest's packageManager field; .yarnrc.yml
# carries the supply-chain age gate into this install.
COPY --from=builder /daytona/dist/apps/api/package.json /daytona/dist/apps/api/yarn.lock ./
COPY --from=builder /daytona/.yarnrc.yml ./

RUN yarn install --immutable

COPY --from=builder /daytona/dist ./dist

# The migration:run:* scripts execute the typeorm CLI through ts-node against
# the TypeScript data-sources (apps/api/src/migrations/**), which load entities
# via a `*.entity.ts` glob. Ship that source + the tsconfigs they extend so the
# deploy migration jobs resolve identically to a full checkout.
COPY --from=builder /daytona/apps/api/src ./apps/api/src
COPY --from=builder /daytona/apps/api/tsconfig.json /daytona/apps/api/tsconfig.app.json /daytona/apps/api/tsconfig.spec.json ./apps/api/
COPY --from=builder /daytona/tsconfig.base.json ./

ARG VERSION=0.0.1
ENV VERSION=${VERSION}

HEALTHCHECK CMD [ "curl", "-f", "http://localhost:3000/api/config" ]

ENTRYPOINT ["node", "dist/apps/api/main.js"]
