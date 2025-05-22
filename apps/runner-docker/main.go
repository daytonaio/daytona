package main

import (
	"log"
	"net"

	"github.com/daytonaio/runner/apps/runner-docker/server"
	pb "github.com/daytonaio/runner/proto"
	"google.golang.org/grpc"
)

func main() {
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	s := grpc.NewServer()
	pb.RegisterRunnerServer(s, server.NewRunnerServer())

	log.Printf("Server listening at %v", lis.Addr())
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
