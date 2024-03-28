# Copyright 2024 Daytona Platforms Inc.
# SPDX-License-Identifier: Apache-2.0

FROM alpine:3.19.1

ARG TARGETOS TARGETARCH
ARG DAYTONA_VERSION
ARG DAYTONA_DIST=daytona-${TARGETOS}-${TARGETARCH}

LABEL \
	org.opencontainers.image.description="The Open Source Development Environment Manager" \
	org.opencontainers.image.title="Daytona" \
	org.opencontainers.image.url="https://github.com/daytonaio/daytona" \
	org.opencontainers.image.vendor="Daytona Platforms Inc." \
	org.opencontainers.image.version="$DAYTONA_VERSION"

RUN addgroup \
		-g 1000 \
		daytona

RUN adduser \
		-D \
		-s /bin/bash \
		-h /home/daytona \
		-u 1000 \
		-G daytona \
		daytona

RUN apk add --no-cache \
		curl \
        gcompat

COPY --chown=daytona:daytona --chmod=755 $DAYTONA_DIST /usr/local/bin/daytona

USER daytona
ENV HOME=/home/daytona
WORKDIR /home/daytona

ENTRYPOINT [ "/usr/local/bin/daytona", "server" ]