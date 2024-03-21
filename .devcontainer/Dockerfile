FROM ubuntu:22.04

ADD postcreate.sh /setup/postcreate.sh
RUN chmod +x /setup/postcreate.sh

RUN apt update -y && \
    apt install curl protobuf-compiler openjdk-11-jdk libprotobuf-dev -y

RUN curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin v1.57.1