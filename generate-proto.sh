#!/bin/bash
protoc \
  --experimental_allow_proto3_optional \
  --plugin=./node_modules/.bin/protoc-gen-ts_proto \
  --ts_proto_out=./libs/runner-grpc-client/src \
  --ts_proto_opt=useOptionals=messages \
  --ts_proto_opt=forceLong=false \
  --ts_proto_opt=outputServices=nice-grpc \
  --ts_proto_opt=esModuleInterop=true \
  --proto_path=./apps/proto \
  ./apps/proto/*.proto

  protoc \
    --go_out=. \
    --go-grpc_out=. \
    --go_opt=paths=source_relative \
    --go-grpc_opt=paths=source_relative \
    --experimental_allow_proto3_optional \
    apps/proto/runner.proto
