npm install -g @devcontainers/cli

go get -u google.golang.org/protobuf/cmd/protoc-gen-go
go install google.golang.org/protobuf/cmd/protoc-gen-go

go get -u google.golang.org/grpc/cmd/protoc-gen-go-grpc
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc

go mod tidy

echo 'alias daytona="go run cmd/daytona/main.go"' >> ~/.zshrc