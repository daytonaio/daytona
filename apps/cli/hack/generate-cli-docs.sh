#!/bin/bash
# Copyright 2025 Daytona Platforms Inc.
# SPDX-License-Identifier: AGPL-3.0


# Clean up existing documentation files
rm -rf docs hack/docs

# Generate default CLI documentation files in folder "docs"
go run main.go generate-docs

# Normalize the generated YAML to the repo prettier style so that
# 'yarn format' produces no diff in CI (see the generate-api-clients job)
npx prettier --write "hack/docs/**/*.yaml" --log-level warn