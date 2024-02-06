# Copyright 2024 Daytona Platforms Inc.
# SPDX-License-Identifier: Apache-2.0

groupadd -f docker
usermod -aG docker daytona
newgrp docker