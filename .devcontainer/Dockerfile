FROM buildpack-deps:noble-curl

ARG TARGETARCH

# common tools
RUN apt update && export DEBIAN_FRONTEND=noninteractive \
    && apt -y install --no-install-recommends apt-utils vim htop telnet socat expect-dev tini psmisc libgit2-dev \
    python3 python3-pip libx11-dev libxtst-dev libxext-dev libxrandr-dev libxinerama-dev libxi-dev \
    libx11-6 libxrandr2 libxext6 libxrender1 libxfixes3 libxss1 libxtst6 libxi6 \
    xvfb x11vnc novnc xfce4 xfce4-terminal dbus-x11

# build tools
RUN apt update && export DEBIAN_FRONTEND=noninteractive \
    && apt -y install --no-install-recommends openjdk-11-jdk protobuf-compiler libprotobuf-dev

# Disable default sources \
RUN mv /etc/apt/sources.list.d/ubuntu.sources /etc/apt/sources.list.d/ubuntu.sources.disabled

# Copy multi-arch sources
COPY ubuntu.sources /etc/apt/sources.list.d/ubuntu.sources

# Install cross-compilation dependencies
RUN if [ "$TARGETARCH" = "amd64" ]; then \
    CROSS_ARCH="arm64"; \
    CROSS_COMPILER="aarch64-linux-gnu"; \
else \
    CROSS_ARCH="amd64"; \
    CROSS_COMPILER="x86-64-linux-gnu"; \
fi && \
dpkg --add-architecture ${CROSS_ARCH} && \
apt-get update && \
export DEBIAN_FRONTEND=noninteractive && \
apt-get install -y \
    gcc-${CROSS_COMPILER} \
    g++-${CROSS_COMPILER} \
    libc6-dev:${CROSS_ARCH} \
    libx11-dev:${CROSS_ARCH} \
    libxtst-dev:${CROSS_ARCH} \
    libxkbcommon-dev:${CROSS_ARCH} \
    libpng-dev:${CROSS_ARCH} \
    zlib1g-dev:${CROSS_ARCH} && \
rm -rf /var/lib/apt/lists/*

# Reclaim UID 1000 by reassigning 'ubuntu' user
RUN usermod -u 2000 ubuntu && groupmod -g 2000 ubuntu

CMD ["tail", "-f", "/dev/null"]
