#!/bin/bash

set -e  # Exit on error
set -x  # Print commands as they are executed

echo "Starting proto generation..."

# Check if protoc is installed
if ! command -v protoc &> /dev/null; then
    echo "Error: protoc is not installed"
    exit 1
fi

# Check if proto file exists
if [ ! -f "runner.proto" ]; then
    echo "Error: runner.proto not found in current directory"
    exit 1
fi

echo "Generating TypeScript client..."
# Generate TypeScript client
protoc \
  --experimental_allow_proto3_optional \
  --plugin=../../node_modules/.bin/protoc-gen-ts_proto \
  --ts_proto_out=../../libs/runner-grpc-client/src \
  --ts_proto_opt=useOptionals=messages \
  --ts_proto_opt=forceLong=false \
  --ts_proto_opt=outputServices=nice-grpc \
  --ts_proto_opt=esModuleInterop=true \
  --proto_path=. \
  ./*.proto

echo "Generating Go client..."
# Generate Go client
protoc \
  --go_out=../runner \
  --go-grpc_out=../runner \
  --go_opt=paths=source_relative \
  --go-grpc_opt=paths=source_relative \
  --experimental_allow_proto3_optional \
  runner.proto

echo "Proto generation completed successfully"

ls -l ../../node_modules/.bin/protoc-gen-ts_proto 

protoc --version 
