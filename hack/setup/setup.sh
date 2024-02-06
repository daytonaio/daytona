# Copyright 2024 Daytona Platforms Inc.
# SPDX-License-Identifier: Apache-2.0

DEBIAN_FRONTEND=noninteractive apt update -y && \
DEBIAN_FRONTEND=noninteractive apt install docker.io curl -y

curl https://daytona-demo.s3.amazonaws.com/daytona --output daytona
chmod +x daytona
mv daytona /usr/local/bin