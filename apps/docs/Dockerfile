FROM node:22-alpine AS build

ENV CI=true

RUN npm install -g corepack && corepack enable

WORKDIR /daytona

# Yarn caching layer
COPY package.json yarn.lock .yarnrc.yml ./
RUN yarn install --immutable

# Nx + TS config
COPY nx.json tsconfig.base.json ./

# App source (docs + api for openapi generation)
COPY apps/docs/ apps/docs/
COPY apps/api/ apps/api/
COPY apps/daemon/pkg/toolbox/docs/swagger.json apps/daemon/pkg/toolbox/docs/swagger.json

# Lib dependencies (api imports runner-api-client)
COPY libs/runner-api-client/ libs/runner-api-client/
COPY libs/runner-proto/ libs/runner-proto/

# Docs build arguments
ENV PUBLIC_WEB_URL=https://daytona.io
ARG PUBLIC_ALGOLIA_APP_ID
ENV PUBLIC_ALGOLIA_APP_ID=${PUBLIC_ALGOLIA_APP_ID}
ARG PUBLIC_ALGOLIA_API_KEY
ENV PUBLIC_ALGOLIA_API_KEY=${PUBLIC_ALGOLIA_API_KEY}
ARG PUBLIC_WEB_URL
ENV PUBLIC_WEB_URL=${PUBLIC_WEB_URL}
ARG PUBLIC_ALGOLIA_DOCS_INDEX_NAME=docs
ENV PUBLIC_ALGOLIA_DOCS_INDEX_NAME=${PUBLIC_ALGOLIA_DOCS_INDEX_NAME}
ARG PUBLIC_ALGOLIA_CLI_INDEX_NAME=cli
ENV PUBLIC_ALGOLIA_CLI_INDEX_NAME=${PUBLIC_ALGOLIA_CLI_INDEX_NAME}
ARG PUBLIC_ALGOLIA_SDK_INDEX_NAME=sdk
ENV PUBLIC_ALGOLIA_SDK_INDEX_NAME=${PUBLIC_ALGOLIA_SDK_INDEX_NAME}
ARG PUBLIC_DOCSEARCH_APP_ID
ENV PUBLIC_DOCSEARCH_APP_ID=${PUBLIC_DOCSEARCH_APP_ID}
ARG PUBLIC_DOCSEARCH_API_KEY
ENV PUBLIC_DOCSEARCH_API_KEY=${PUBLIC_DOCSEARCH_API_KEY}
ARG PUBLIC_DOCSEARCH_INDEX_NAME=ai
ENV PUBLIC_DOCSEARCH_INDEX_NAME=${PUBLIC_DOCSEARCH_INDEX_NAME}
ARG PUBLIC_DOCSEARCH_ASSISTANT_ID
ENV PUBLIC_DOCSEARCH_ASSISTANT_ID=${PUBLIC_DOCSEARCH_ASSISTANT_ID}

ENV VERSION=12345.67890.0
ENV NX_DAEMON=false

RUN yarn nx build docs --configuration=production --nxBail=true

FROM node:22-alpine AS docs

ARG VERSION=0.0.1

WORKDIR /daytona

COPY --from=build /daytona/node_modules node_modules
COPY --from=build /daytona/dist/apps/docs dist/apps/docs
COPY --from=build /daytona/apps/docs/server dist/apps/docs/server

WORKDIR /daytona/dist/apps/docs

RUN VER=$(echo ${VERSION} | sed 's/^v//') && \
  MAJOR=$(echo ${VER} | cut -d. -f1) && \
  MINOR=$(echo ${VER} | cut -d. -f2) && \
  find . -type f \( -name "*.mjs" -o -name "*.js" -o -name "*.html" \) \
    -exec sed -i "s/12345\.67890/${MAJOR}.${MINOR}/g" {} +

ENTRYPOINT ["sh", "-c", "node server/index.mjs"]
