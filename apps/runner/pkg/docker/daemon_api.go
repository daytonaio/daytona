// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package docker

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/daytonaio/runner/cmd/runner/config"
)

type apiAccess struct {
	ApiUrl string `json:"apiUrl"`
	Token  string `json:"token"`
}

func (d *DockerClient) configureDaemonApiAccess(containerIP string, token string) error {
	cfg, err := config.GetConfig()
	if err != nil {
		return fmt.Errorf("failed to get config: %w", err)
	}

	access := apiAccess{
		ApiUrl: cfg.ServerUrl,
		Token:  token,
	}

	jsonData, err := json.Marshal(access)
	if err != nil {
		return fmt.Errorf("failed to marshal API access data: %w", err)
	}

	url := fmt.Sprintf("http://%s:2280/api-access", containerIP)
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to send API access to daemon: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("daemon returned non-200 status code: %d", resp.StatusCode)
	}

	return nil
}
