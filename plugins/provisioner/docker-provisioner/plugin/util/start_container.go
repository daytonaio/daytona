package util

import (
	"archive/tar"
	"bytes"
	"context"
	"io"
	"os"
	"time"

	"github.com/daytonaio/daytona/agent/workspace"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	log "github.com/sirupsen/logrus"
)

func StartContainer(project workspace.Project) error {
	containerName := GetContainerName(project)
	ctx := context.Background()

	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return err
	}

	err = cli.ContainerStart(ctx, containerName, types.ContainerStartOptions{})
	if err != nil {
		return err
	}

	// make sure container is running
	//	TODO: timeout
	for {
		inspect, err := cli.ContainerInspect(ctx, containerName)
		if err != nil {
			return err
		}

		if inspect.State.Running {
			break
		}

		time.Sleep(1 * time.Second)
	}

	//	copy daytona binary
	daytonaPath, err := os.Executable()
	if err != nil {
		return err
	}

	file, err := os.Open(daytonaPath)
	if err != nil {
		return err
	}
	defer file.Close()

	buf := new(bytes.Buffer)
	tw := tar.NewWriter(buf)
	fileInfo, err := file.Stat()
	if err != nil {
		return err
	}

	header, err := tar.FileInfoHeader(fileInfo, fileInfo.Name())
	if err != nil {
		return err
	}

	// Set the name of the file in the tar archive
	header.Name = "daytona"

	err = tw.WriteHeader(header)
	if err != nil {
		return err
	}

	_, err = io.Copy(tw, file)
	if err != nil {
		return err
	}

	err = tw.Close()
	if err != nil {
		return err
	}

	// Copy the file content to the container
	err = cli.CopyToContainer(ctx, containerName, "/usr/local/bin", buf, types.CopyToContainerOptions{})
	if err != nil {
		return err
	}

	// start dockerd
	execConfig := types.ExecConfig{
		Tty:          true,
		AttachStdout: true,
		AttachStderr: true,
		Cmd:          []string{"dockerd"},
		User:         "root",
	}
	execResp, err := cli.ContainerExecCreate(ctx, containerName, execConfig)
	if err != nil {
		return err
	}

	log.Debug("Daytona binary copied to container")

	err = cli.ContainerExecStart(ctx, execResp.ID, types.ExecStartCheck{})
	if err != nil {
		return err
	}

	//	todo: wait for dockerd to start
	time.Sleep(3 * time.Second)

	return nil
}
