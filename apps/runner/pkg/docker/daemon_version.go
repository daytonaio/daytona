// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package docker

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/daytonaio/runner/pkg/common"
)

type daemonVersionResponse struct {
	Version string `json:"version" example:"0.1.0"`
}

func (d *DockerClient) GetDaemonVersion(ctx context.Context, sandboxId string) (string, error) {
	c, err := d.ContainerInspect(ctx, sandboxId)
	if err != nil {
		return "", err
	}

	containerIP := common.GetContainerIpAddress(ctx, c)
	if containerIP == "" {
		return "", errors.New("sandbox IP not found? Is the sandbox started?")
	}

	targetUrl := fmt.Sprintf("http://%s:2280/version", containerIP)
	target, err := url.Parse(targetUrl)
	if err != nil {
		return "", err
	}

	return d.getDaemonVersion(ctx, target)
}

func (d *DockerClient) getDaemonVersion(ctx context.Context, targetUrl *url.URL) (string, error) {
	client := http.Client{
		Timeout: 1 * time.Second,
	}
	resp, err := client.Get(targetUrl.String())
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var versionResponse daemonVersionResponse
	err = json.NewDecoder(resp.Body).Decode(&versionResponse)
	if err != nil {
		return "", err
	}

	return versionResponse.Version, nil
}
