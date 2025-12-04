# syntax=docker/dockerfile:1

FROM golang:1.25-bookworm AS build

WORKDIR /app

# Copy go mod files
COPY apps/mock-runner/go.mod apps/mock-runner/go.sum ./apps/mock-runner/
COPY libs/common-go/go.mod libs/common-go/go.sum ./libs/common-go/

# Download dependencies
RUN cd apps/mock-runner && go mod download

# Copy source code
COPY apps/mock-runner ./apps/mock-runner
COPY libs/common-go ./libs/common-go

# Build the application
ARG VERSION=v0.0.0-dev
RUN cd apps/mock-runner && \
    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags "-X 'github.com/daytonaio/mock-runner/internal.Version=${VERSION}'" \
    -o /mock-runner \
    ./cmd/mock-runner/main.go

# Final stage
FROM debian:bookworm-slim

# Install ca-certificates for HTTPS
RUN apt-get update && apt-get install -y ca-certificates && rm -rf /var/lib/apt/lists/*

WORKDIR /app

# Copy the binary
COPY --from=build /mock-runner /app/mock-runner

# Copy daemon binary if available (will be added during build process)
COPY --from=build /app/apps/mock-runner/pkg/daemon/static/daemon-amd64 /app/daemon-amd64 2>/dev/null || true

# Expose API port
EXPOSE 8080

# Run the mock runner
ENTRYPOINT ["/app/mock-runner"]



