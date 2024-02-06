# Copyright 2024 Daytona Platforms Inc.
# SPDX-License-Identifier: Apache-2.0

# Generates the configuration.json file for the code-server setup in the workspace project container.
# Generated file content must be manually updated in the internal/scripts/server/configuration.go file

devcontainer read-configuration --include-merged-configuration --log-format json --workspace-folder=. |& tee code-server/configuration.json

# echo $(tail -n 1 code-server/configuration.json) > code-server/configuration.json