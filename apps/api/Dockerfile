FROM node:24-slim AS daytona

# Install dependencies (apt instead of apk)
RUN apt-get update && apt-get install -y --no-install-recommends bash curl && \
  rm -rf /var/lib/apt/lists/*
RUN npm install -g corepack && corepack enable

WORKDIR /daytona

COPY . .

ARG VERSION=0.0.1
ENV VERSION=${VERSION}

RUN yarn

RUN yarn nx build api --configuration=production --nxBail=true && \
  VITE_BASE_API_URL=%DAYTONA_BASE_API_URL% yarn nx build dashboard --configuration=production --nxBail=true

HEALTHCHECK CMD [ "curl", "-f", "http://localhost:3000/api/config" ]

ENTRYPOINT ["node", "dist/apps/api/main.js"]