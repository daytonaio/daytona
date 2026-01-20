# Copyright 2025 Daytona Platforms Inc.
# SPDX-License-Identifier: AGPL-3.0

# Build stage
FROM golang:1.23-bookworm AS builder

WORKDIR /app

# Install build dependencies
RUN apt-get update && apt-get install -y \
    git \
    ca-certificates \
    && rm -rf /var/lib/apt/lists/*

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the binary
ARG VERSION=0.0.0-dev
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags "-X 'github.com/daytonaio/runner-ch/internal.Version=${VERSION}'" \
    -o /runner-ch ./cmd/runner

# Runtime stage
FROM ubuntu:24.04 AS runner

# Install runtime dependencies
RUN apt-get update && apt-get install -y \
    ca-certificates \
    openssh-client \
    socat \
    iproute2 \
    iptables \
    qemu-utils \
    wget \
    curl \
    && rm -rf /var/lib/apt/lists/*

# Install Cloud Hypervisor (for local mode)
ARG CH_VERSION=v50.0.0
RUN wget -q https://github.com/cloud-hypervisor/cloud-hypervisor/releases/download/${CH_VERSION}/cloud-hypervisor-static \
    -O /usr/local/bin/cloud-hypervisor \
    && chmod +x /usr/local/bin/cloud-hypervisor

# Install ch-remote
RUN wget -q https://github.com/cloud-hypervisor/cloud-hypervisor/releases/download/${CH_VERSION}/ch-remote-static \
    -O /usr/local/bin/ch-remote \
    && chmod +x /usr/local/bin/ch-remote

# Create runtime directories
RUN mkdir -p /var/lib/cloud-hypervisor/{sandboxes,snapshots,images,kernels,firmware} \
    && mkdir -p /var/run/cloud-hypervisor

# Create TAP helper scripts
RUN echo '#!/bin/bash\nTAP_NAME=$1\nip tuntap add $TAP_NAME mode tap\nip link set $TAP_NAME master ${CH_BRIDGE_NAME:-br0}\nip link set $TAP_NAME up' > /usr/local/bin/ch-create-tap \
    && chmod +x /usr/local/bin/ch-create-tap

RUN echo '#!/bin/bash\nTAP_NAME=$1\nip link del $TAP_NAME 2>/dev/null || true' > /usr/local/bin/ch-delete-tap \
    && chmod +x /usr/local/bin/ch-delete-tap

# Copy the runner binary
COPY --from=builder /runner-ch /usr/local/bin/runner-ch

# Copy daemon binary (will be embedded during build)
COPY pkg/daemon/static/ /opt/daytona/

# Set environment variables
ENV API_PORT=8080
ENV CH_SOCKETS_PATH=/var/run/cloud-hypervisor
ENV CH_SANDBOXES_PATH=/var/lib/cloud-hypervisor/sandboxes
ENV CH_SNAPSHOTS_PATH=/var/lib/cloud-hypervisor/snapshots
ENV CH_KERNEL_PATH=/var/lib/cloud-hypervisor/kernels/vmlinuz-6.8.0-90-generic
ENV CH_INITRAMFS_PATH=/var/lib/cloud-hypervisor/kernels/initrd.img-6.8.0-90-generic
ENV CH_FIRMWARE_PATH=/var/lib/cloud-hypervisor/firmware/hypervisor-fw
ENV CH_BASE_IMAGE_PATH=/var/lib/cloud-hypervisor/snapshots/ubuntu-base.1.qcow2
ENV CH_BRIDGE_NAME=br0

# Expose API port
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD curl -f http://localhost:${API_PORT}/ || exit 1

# Run the runner
ENTRYPOINT ["/usr/local/bin/runner-ch"]
