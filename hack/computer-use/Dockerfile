FROM ubuntu:22.04

COPY --from=golang:1.23 /usr/local/go /usr/local/go
ENV PATH="/usr/local/go/bin:${PATH}"

WORKDIR /app 
COPY ./libs/computer-use /app/libs/computer-use
COPY ./libs/api-client-go /app/libs/api-client-go
COPY ./libs/common-go /app/libs/common-go
COPY ./apps/runner /app/apps/runner
COPY ./apps/cli /app/apps/cli
COPY ./apps/proxy /app/apps/proxy
COPY ./apps/daemon /app/apps/daemon
COPY go.work /app/go.work
COPY go.work.sum /app/go.sum

ENV DEBIAN_FRONTEND=noninteractive

RUN apt-get update && apt-get install -y ca-certificates && update-ca-certificates
RUN apt-get install -y libx11-dev libxtst-dev gcc

RUN go build -o computer-use /app/libs/computer-use/main.go

RUN chmod +x /app/computer-use

VOLUME ["/dist"]

ENTRYPOINT ["cp", "/app/computer-use", "/dist/libs/computer-use-amd64"]