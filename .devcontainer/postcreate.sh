#!/bin/bash

go install github.com/swaggo/swag/cmd/swag@v1.16.3

go mod tidy

echo 'alias dtn="DAYTONA_DEV=1 go run cmd/daytona/main.go"' >> ~/.zshrc