package server

import (
	pb "github.com/daytonaio/runner/proto"
)

type RunnerServer struct {
	pb.UnimplementedRunnerServer
}

func NewRunnerServer() *RunnerServer {
	return &RunnerServer{}
}
