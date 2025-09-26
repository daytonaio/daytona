FROM node:22-alpine AS build

RUN npm install -g corepack && corepack enable

WORKDIR /daytona

COPY . .

# Docs build arguments
ENV PUBLIC_WEB_URL=https://daytona.io
ARG PUBLIC_ALGOLIA_APP_ID
ENV PUBLIC_ALGOLIA_APP_ID=${PUBLIC_ALGOLIA_APP_ID}
ARG PUBLIC_ALGOLIA_API_KEY
ENV PUBLIC_ALGOLIA_API_KEY=${PUBLIC_ALGOLIA_API_KEY}
ARG PUBLIC_WEB_URL
ENV PUBLIC_WEB_URL=${PUBLIC_WEB_URL}

ARG VERSION=0.0.1
ENV VERSION=${VERSION}

RUN yarn

RUN yarn nx build docs --configuration=production --nxBail=true

FROM node:22-alpine AS docs

WORKDIR /daytona

COPY --from=build /daytona/node_modules node_modules
COPY --from=build /daytona/dist/apps/docs dist/apps/docs
COPY --from=build /daytona/apps/docs/server dist/apps/docs/server

WORKDIR /daytona/dist/apps/docs

ENTRYPOINT ["sh", "-c", "node server/index.mjs"]