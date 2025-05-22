package server

import (
	pb "github.com/daytonaio/runner/proto"
)

func (s *RunnerServer) GetBuildLogs(req *pb.GetBuildLogsRequest, stream pb.Runner_GetBuildLogsServer) error {
	// TODO: Implement build logs streaming logic
	return stream.Send(&pb.LogLine{
		Content: "Build completed successfully",
	})
}
