// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package server

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/daytonaio/runner-docker/internal/constants"
	"github.com/daytonaio/runner-docker/internal/util"
	"github.com/daytonaio/runner-docker/pkg/models/enums"
	pb "github.com/daytonaio/runner/proto"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/registry"
	"github.com/docker/docker/pkg/jsonmessage"
	log "github.com/sirupsen/logrus"
)

func (s *RunnerServer) PullImage(ctx context.Context, req *pb.PullImageRequest) (*pb.PullImageResponse, error) {
	tag := "latest"
	lastColonIndex := strings.LastIndex(req.Image, ":")
	if lastColonIndex != -1 {
		tag = req.Image[lastColonIndex+1:]
	}

	if tag != "latest" {
		resp, err := s.ImageExists(ctx, &pb.ImageExistsRequest{
			Image:         req.Image,
			IncludeLatest: true,
		})
		if err != nil {
			return nil, err
		}

		if resp.Exists {
			return &pb.PullImageResponse{
				Message: fmt.Sprintf("Image %s already pulled", req.Image),
			}, nil
		}
	}

	log.Infof("Pulling image %s...", req.Image)

	sandboxIdValue := ctx.Value(constants.ID_KEY)

	if sandboxIdValue != nil {
		sandboxId := sandboxIdValue.(string)
		s.cache.SetSandboxState(ctx, sandboxId, enums.SandboxStatePullingImage)
	}

	responseBody, err := s.dockerClient.ImagePull(ctx, req.Image, image.PullOptions{
		RegistryAuth: getRegistryAuth(req.Registry),
	})
	if err != nil {
		return nil, err
	}
	defer responseBody.Close()

	err = jsonmessage.DisplayJSONMessagesStream(responseBody, io.Writer(&util.DebugLogWriter{}), 0, true, nil)
	if err != nil {
		return nil, err
	}

	log.Infof("Image %s pulled successfully", req.Image)

	return &pb.PullImageResponse{
		Message: fmt.Sprintf("Image %s pulled", req.Image),
	}, nil
}

func getRegistryAuth(reg *pb.Registry) string {
	if reg == nil {
		// Sometimes registry auth fails if "" is sent, so sending "empty" instead
		return "empty"
	}

	authConfig := registry.AuthConfig{
		// TODO: nil-check
		Username: *reg.Username,
		Password: *reg.Password,
	}
	encodedJSON, err := json.Marshal(authConfig)
	if err != nil {
		// Sometimes registry auth fails if "" is sent, so sending "empty" instead
		return "empty"
	}

	return base64.URLEncoding.EncodeToString(encodedJSON)
}
