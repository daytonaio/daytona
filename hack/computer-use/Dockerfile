FROM ubuntu:22.04

# Set environment variables early
ENV DEBIAN_FRONTEND=noninteractive

# Install Go and system dependencies in optimized layers
COPY --from=golang:1.23 /usr/local/go /usr/local/go
ENV PATH="/usr/local/go/bin:${PATH}"

# Install all system dependencies in a single layer for better caching
RUN apt-get update && apt-get install -y \
    ca-certificates \
    libx11-dev \
    libxtst-dev \
    gcc \
    && update-ca-certificates \
    && apt-get clean \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /app

RUN echo "go 1.25.4\n\nuse (\n  ./libs/computer-use\n   ./apps/daemon\n     ./libs/common-go\n)" > /app/go.work

COPY go.work.sum /app/go.work.sum

COPY ./apps/daemon /app/apps/daemon
COPY ./libs/common-go /app/libs/common-go
COPY ./libs/computer-use /app/libs/computer-use

# Build the application with optimizations
RUN cd /app/libs/computer-use && \
    CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o /app/computer-use main.go && \
    chmod +x /app/computer-use

VOLUME ["/dist"]

ENTRYPOINT ["cp", "/app/computer-use", "/dist/libs/computer-use-amd64"]
