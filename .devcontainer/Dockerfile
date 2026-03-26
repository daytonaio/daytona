FROM buildpack-deps:jammy-curl

ARG TARGETARCH

# common tools
RUN apt update && export DEBIAN_FRONTEND=noninteractive \
    && apt -y install --no-install-recommends apt-utils vim htop telnet socat expect-dev tini psmisc libgit2-dev \
    python3 python3-pip libx11-dev libxtst-dev libxext-dev libxrandr-dev libxinerama-dev libxi-dev \
    libx11-6 libxrandr2 libxext6 libxrender1 libxfixes3 libxss1 libxtst6 libxi6 \
    xvfb x11vnc novnc xfce4 xfce4-terminal dbus-x11 dnsmasq gettext-base

# build tools
RUN apt update && export DEBIAN_FRONTEND=noninteractive \
    && apt -y install --no-install-recommends openjdk-11-jdk protobuf-compiler libprotobuf-dev

# Mount-s3 (AWS S3 FUSE driver)
RUN /bin/bash -c 'apt update && export DEBIAN_FRONTEND=noninteractive \
    && apt -y install --no-install-recommends fuse libfuse2 \
    && MOUNT_S3_ARCH=$([ "$TARGETARCH" = "amd64" ] && echo "x86_64" || echo "arm64") \
    && wget https://s3.amazonaws.com/mountpoint-s3-release/1.20.0/${MOUNT_S3_ARCH}/mount-s3-1.20.0-${MOUNT_S3_ARCH}.deb \
    && apt install -y ./mount-s3-1.20.0-${MOUNT_S3_ARCH}.deb \
    && rm -f mount-s3-1.20.0-${MOUNT_S3_ARCH}.deb'
# Telepresence
RUN curl -fL https://app.getambassador.io/download/tel2oss/releases/download/v2.17.0/telepresence-linux-${TARGETARCH} -o /usr/local/bin/telepresence && \
    chmod a+x /usr/local/bin/telepresence

RUN echo "address=/proxy.localhost/127.0.0.1" >> /etc/dnsmasq.conf

CMD ["sh", "-c", "service dnsmasq start && echo 'nameserver 127.0.0.1' > /etc/resolv.conf && tail -f /dev/null"]