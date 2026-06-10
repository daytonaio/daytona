/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

package common

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/daytonaio/daytona/cli/config"
	"github.com/daytonaio/daytona/cli/internal/clierr"
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

// ReadBuildLogs streams build logs to stdout until the stream ends or ctx is
// canceled. Cancellation is the caller's signal to stop and returns nil; in
// follow mode an EOF keeps polling for more output until then.
func ReadBuildLogs(ctx context.Context, params ReadLogParams) error {
	url := fmt.Sprintf("%s/%s/%s/build-logs", params.ServerUrl, params.ResourceType, params.Id)
	if params.Follow != nil && *params.Follow {
		url = fmt.Sprintf("%s?follow=true", url)
	}

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return clierr.Newf(clierr.CategoryNetwork, "failed to create request: %v", err)
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
		if ctx.Err() != nil {
			return nil
		}
		return clierr.Newf(clierr.CategoryNetwork, "failed to connect to server: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return clierr.FromHTTPStatus(resp.StatusCode, fmt.Sprintf("server returned a non-OK status while retrieving logs: %d", resp.StatusCode))
	}

	reader := bufio.NewReader(resp.Body)
	buffer := make([]byte, 4096)

	for {
		select {
		case <-ctx.Done():
			return nil
		default:
			n, err := reader.Read(buffer)
			if n > 0 {
				fmt.Print(string(buffer[:n]))
			}

			if err != nil {
				if errors.Is(err, io.EOF) {
					if params.Follow != nil && *params.Follow {
						time.Sleep(500 * time.Millisecond)
						continue
					}
					return nil
				}
				if ctx.Err() != nil || errors.Is(err, context.Canceled) {
					return nil
				}
				// In follow mode the upstream proxy can reset long-lived streams; the
				// build itself is unaffected and the caller still polls for completion.
				if params.Follow != nil && *params.Follow {
					log.Warnf("Build log stream interrupted (%v); the build is still running, waiting for it to complete...", err)
					return nil
				}
				return clierr.Newf(clierr.CategoryNetwork, "error reading from stream: %v", err)
			}
		}
	}
}
