# Copyright 2024 Daytona Platforms Inc.
# SPDX-License-Identifier: Apache-2.0

FROM ubuntu:22.04

RUN apt update -y && \
    apt install curl libgit2-1.1 -y

# Install code-server to skip the setup on workspace initialization
ADD code-server/setup.sh /setup/server/setup.sh
RUN chmod +x /setup/server/setup.sh
ADD code-server/configuration.json /setup/server/configuration.json
# ADD docker/postinstall.sh /setup/docker/postinstall.sh
# RUN chmod +x /setup/docker/postinstall.sh

CMD ["tail", "-f", "/dev/null"]