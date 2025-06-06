/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

package common

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/daytonaio/daytona/cli/config"
	log "github.com/sirupsen/logrus"
)

type ReadLogParams struct {
	Id                   string
	ServerUrl            string
	ServerApi            config.ServerApi
	ActiveOrganizationId *string
	Follow               *bool
	ResourceType         ResourceType
}

type ResourceType string

const (
	ResourceTypeSandbox  ResourceType = "sandbox"
	ResourceTypeSnapshot ResourceType = "snapshots"
)

func ReadBuildLogs(ctx context.Context, params ReadLogParams) {
	url := fmt.Sprintf("%s/%s/%s/build-logs", params.ServerUrl, params.ResourceType, params.Id)
	if params.Follow != nil && *params.Follow {
		url = fmt.Sprintf("%s?follow=true", url)
	}

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		log.Errorf("Failed to create request: %v", err)
		return
	}

	if params.ServerApi.Key != nil {
		req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", *params.ServerApi.Key))
	} else if params.ServerApi.Token != nil {
		req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", params.ServerApi.Token.AccessToken))

		if params.ActiveOrganizationId != nil {
			req.Header.Add("X-Daytona-Organization-ID", *params.ActiveOrganizationId)
		}
	}

	req.Header.Add("Accept", "application/octet-stream")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Errorf("Failed to connect to server: %v", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Errorf("Server returned a non-OK status while retrieving logs: %d", resp.StatusCode)
		return
	}

	reader := bufio.NewReader(resp.Body)
	buffer := make([]byte, 4096)

	for {
		select {
		case <-ctx.Done():
			return
		default:
			n, err := reader.Read(buffer)
			if n > 0 {
				fmt.Print(string(buffer[:n]))
			}

			if err != nil {
				if err == io.EOF {
					if params.Follow != nil && *params.Follow {
						time.Sleep(500 * time.Millisecond)
						continue
					}
					return
				}
				log.Errorf("Error reading from stream: %v", err)
				return
			}
		}
	}
}
