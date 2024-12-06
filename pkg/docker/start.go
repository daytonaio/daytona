// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package docker

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os/exec"
	"time"

	"github.com/daytonaio/daytona/pkg/build/detect"
	"github.com/daytonaio/daytona/pkg/common"
	"github.com/daytonaio/daytona/pkg/models"
	"github.com/daytonaio/daytona/pkg/provider/util"
	"github.com/docker/docker/api/types/container"
)

func (d *DockerClient) StartWorkspace(opts *CreateWorkspaceOptions, daytonaDownloadUrl string) error {
	var err error
	containerUser := opts.Workspace.User

	builderType, err := detect.DetectWorkspaceBuilderType(opts.Workspace.BuildConfig, opts.WorkspaceDir, opts.SshClient)
	if err != nil {
		return err
	}

	switch builderType {
	case detect.BuilderTypeDevcontainer:
		var remoteUser RemoteUser
		remoteUser, err = d.startDevcontainerWorkspace(opts)
		containerUser = string(remoteUser)
	case detect.BuilderTypeImage:
		err = d.startImageWorkspace(opts)
	default:
		return fmt.Errorf("unknown builder type: %s", builderType)
	}

	if err != nil {
		return err
	}

	if len(opts.ContainerRegistries) > 0 {
		err := d.addContainerRegistriesToDockerConfig(opts.Workspace, containerUser, opts.ContainerRegistries)
		if err != nil {
			return err
		}
	}

	return d.startDaytonaAgent(opts.Workspace, containerUser, daytonaDownloadUrl, opts.LogWriter)
}

func (d *DockerClient) startDaytonaAgent(w *models.Workspace, containerUser, daytonaDownloadUrl string, logWriter io.Writer) error {
	errChan := make(chan error)

	go func() {
		result, err := d.ExecSync(d.GetWorkspaceContainerName(w), container.ExecOptions{
			Cmd:          []string{"bash", "-c", util.GetWorkspaceStartScript(daytonaDownloadUrl, w.ApiKey)},
			AttachStdout: true,
			AttachStderr: true,
			User:         containerUser,
		}, logWriter)
		if err != nil {
			errChan <- err
		}

		if result.ExitCode != 0 {
			errChan <- errors.New(result.StdErr)
		}
	}()

	go func() {
		// TODO: Figure out how to check if the agent is running here
		time.Sleep(5 * time.Second)
		errChan <- nil
	}()

	return <-errChan
}

func (d *DockerClient) addContainerRegistriesToDockerConfig(w *models.Workspace, containerUser string, crs common.ContainerRegistries) error {
	containerRegistriesConfigContent := make(map[string]interface{})
	for _, cr := range crs {
		auth := base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s", cr.Username, cr.Password)))
		containerRegistriesConfigContent[cr.Server] = fmt.Sprintf(`{"auth":"%s"}`, auth)
	}

	authsData := map[string]interface{}{
		"auths": containerRegistriesConfigContent,
	}

	containerName := d.GetWorkspaceContainerName(w)
	dockerConfigPath := "/root/.docker/config.json"
	if containerUser != "root" {
		dockerConfigPath = fmt.Sprintf("/home/%s/.docker/config.json", containerUser)
	}

	cmdRead := exec.Command("docker", "exec", "-i", containerName, "cat", dockerConfigPath)
	out, err := cmdRead.Output()
	if err != nil {
		return fmt.Errorf("failed to read config.json: %v", err)
	}

	// Parse the existing JSON
	var config map[string]interface{}
	if err := json.Unmarshal(out, &config); err != nil {
		return fmt.Errorf("failed to parse JSON: %v", err)
	}

	// Merge the new data into the config
	for key, value := range authsData {
		config[key] = value
	}

	// Convert the updated JSON back to bytes
	updatedConfig, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal updated JSON: %v", err)
	}

	// Write the updated JSON back to the file in the container
	cmdWrite := exec.Command("docker", "exec", "-i", containerName, "sh", "-c", fmt.Sprintf("cat > %s", dockerConfigPath))
	cmdWrite.Stdin = bytes.NewReader(updatedConfig)
	if err := cmdWrite.Run(); err != nil {
		return fmt.Errorf("failed to write updated config.json: %v", err)
	}

	return nil
}
