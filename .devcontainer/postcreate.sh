#!/bin/bash

echo 'alias dtn="DAYTONA_DEV=1 DAYTONA_CONFIG_DIR=$HOME/.config/daytona-dev go run cmd/daytona/main.go"' >> ~/.zshrc

go install github.com/swaggo/swag/cmd/swag@v1.16.3

go mod tidy