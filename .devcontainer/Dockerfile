FROM ubuntu:22.04

ADD postcreate.sh /setup/postcreate.sh
RUN chmod +x /setup/postcreate.sh

RUN apt update -y && \
    apt install curl protobuf-compiler openjdk-11-jdk libprotobuf-dev -y