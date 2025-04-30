// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package apiclient

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type Client struct {
	baseURL    string
	httpClient *http.Client
}
type Sandbox struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

func (c Client) SandboxList(page, limit int) ([]Sandbox, int, error) {
	url := fmt.Sprintf("%s/sandboxes?page=%d&limit=%d", c.baseURL, page, limit)
	resp, err := c.httpClient.Get(url)
	if err != nil {
		return nil, 0, err
	}
	defer resp.Body.Close()

	var result struct {
		Sandboxes []Sandbox `json:"data"`
		Total     int       `json:"total"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, 0, err
	}

	return result.Sandboxes, result.Total, nil
}
