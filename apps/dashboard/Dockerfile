FROM node:20-alpine AS builder

RUN npm install -g corepack && corepack enable

WORKDIR /daytona

COPY . .

RUN yarn

RUN VITE_BASE_API_URL=%DAYTONA_BASE_API_URL% yarn nx build dashboard --configuration=production --nxBail=true

FROM nginx:alpine as dashboard

COPY --from=builder /daytona/dist/apps/dashboard /usr/share/nginx/html
COPY --from=builder /daytona/apps/dashboard/docker/nginx/default.conf /etc/nginx/conf.d/default.conf
COPY --from=builder /daytona/apps/dashboard/docker/entrypoint.sh /entrypoint.sh

ENV DAYTONA_BASE_API_URL="http://api:3000"

EXPOSE 80

HEALTHCHECK CMD [ "wget", "-q", "--spider", "http://localhost/" ]

ENTRYPOINT ["/entrypoint.sh"]
