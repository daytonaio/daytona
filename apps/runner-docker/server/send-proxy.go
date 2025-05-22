package server

import (
	"context"

	pb "github.com/daytonaio/runner/proto"
)

func (s *RunnerServer) SendProxy(ctx context.Context, req *pb.ProxyRequest) (*pb.ProxyResponse, error) {
	// TODO: Implement proxy logic
	return &pb.ProxyResponse{
		StatusCode: 200,
		Headers:    make(map[string]string),
		Body:       []byte("Proxy request processed"),
	}, nil
}
