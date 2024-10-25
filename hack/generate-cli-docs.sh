# Copyright 2024 Daytona Platforms Inc.
# SPDX-License-Identifier: Apache-2.0

# Generate default CLI documentation files in folder "docs"
go run cmd/daytona/main.go generate-docs

# Generate agent mode documentation files in folder "docs/agent_mode"
DAYTONA_TARGET_ID=foo go run cmd/daytona/main.go generate-docs --directory docs/agent_mode