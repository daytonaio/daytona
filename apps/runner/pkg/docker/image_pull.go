// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package docker

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/daytonaio/common-go/pkg/timer"
	"github.com/daytonaio/runner/internal/constants"
	"github.com/daytonaio/runner/internal/util"
	"github.com/daytonaio/runner/pkg/api/dto"
	"github.com/daytonaio/runner/pkg/models/enums"

	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/registry"
	"github.com/docker/docker/errdefs"
	"github.com/docker/docker/pkg/jsonmessage"

	log "github.com/sirupsen/logrus"

	common_errors "github.com/daytonaio/common-go/pkg/errors"
)

func (d *DockerClient) PullImage(ctx context.Context, imageName string, reg *dto.RegistryDTO, validateArchitecture bool) error {
	defer timer.Timer()()

	tag := "latest"
	lastColonIndex := strings.LastIndex(imageName, ":")
	if lastColonIndex != -1 {
		tag = imageName[lastColonIndex+1:]
	}

	if tag != "latest" {
		exists, err := d.ImageExists(ctx, imageName, true)
		if err != nil {
			return err
		}

		if exists {
			if validateArchitecture {
				return d.validateImageArchitecture(ctx, imageName)
			}

			return nil
		}
	}

	log.Infof("Pulling image %s...", imageName)

	sandboxIdValue := ctx.Value(constants.ID_KEY)

	if sandboxIdValue != nil {
		sandboxId := sandboxIdValue.(string)
		d.statesCache.SetSandboxState(ctx, sandboxId, enums.SandboxStatePullingSnapshot)
	}

	responseBody, err := d.apiClient.ImagePull(ctx, imageName, image.PullOptions{
		RegistryAuth: getRegistryAuth(reg),
		Platform:     "linux/amd64",
	})
	if err != nil {
		return err
	}
	defer responseBody.Close()

	err = jsonmessage.DisplayJSONMessagesStream(responseBody, io.Writer(&util.DebugLogWriter{}), 0, true, nil)
	if err != nil {
		return err
	}

	if validateArchitecture {
		if err = d.validateImageArchitecture(ctx, imageName); err != nil {
			log.Errorf("ERROR: %s.\n", err.Error())
			return err
		}
	}

	log.Infof("Image %s pulled successfully", imageName)

	return nil
}

func getRegistryAuth(reg *dto.RegistryDTO) string {
	if reg == nil {
		// Sometimes registry auth fails if "" is sent, so sending "empty" instead
		return "empty"
	}

	authConfig := registry.AuthConfig{
		Username: reg.Username,
		Password: reg.Password,
	}
	encodedJSON, err := json.Marshal(authConfig)
	if err != nil {
		// Sometimes registry auth fails if "" is sent, so sending "empty" instead
		return "empty"
	}

	return base64.URLEncoding.EncodeToString(encodedJSON)
}

func (p *DockerClient) validateImageArchitecture(ctx context.Context, image string) error {
	defer timer.Timer()()

	inspect, err := p.apiClient.ImageInspect(ctx, image)
	if err != nil {
		if errdefs.IsNotFound(err) {
			return err
		}
		return fmt.Errorf("failed to inspect image: %w", err)
	}

	arch := strings.ToLower(inspect.Architecture)
	validArchs := []string{"amd64", "x86_64"}

	for _, validArch := range validArchs {
		if arch == validArch {
			return nil
		}
	}

	return common_errors.NewConflictError(fmt.Errorf("image %s architecture (%s) is not x64 compatible", image, inspect.Architecture))
}
