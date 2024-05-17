# Copyright 2024 Daytona Platforms Inc.
# SPDX-License-Identifier: Apache-2.0

(cd pkg && swag fmt && swag init --parseDependency --parseInternal --parseDepth 1 -o api/docs -g api/server.go)
GO_POST_PROCESS_FILE="/usr/local/bin/gofmt -w" GIT_USER_ID=daytonaio GIT_REPO_ID=daytona npx --yes @openapitools/openapi-generator-cli generate -i pkg/api/docs/swagger.json -g go --package-name=apiclient --additional-properties=isGoSubmodule=true -o pkg/apiclient && rm -rf pkg/apiclient/.openapi-generator/FILES