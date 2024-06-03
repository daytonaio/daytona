# Copyright 2024 Daytona Platforms Inc.
# SPDX-License-Identifier: Apache-2.0

# Generate default CLI documentation files in folder "docs"
go run cmd/daytona/main.go generate-docs

# Generate workspace mode documentation files in folder "docs/workspace_mode"
DAYTONA_WS_ID=foo go run cmd/daytona/main.go generate-docs --directory docs/workspace_mode