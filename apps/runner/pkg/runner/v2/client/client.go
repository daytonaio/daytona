/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

package client

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"

	"github.com/daytonaio/runner/cmd/runner/config"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

type APIClient struct {
	baseURL    string
	token      string
	httpClient *http.Client
}

func NewAPIClient() (*APIClient, error) {
	c, err := config.GetConfig()
	if err != nil {
		return nil, err
	}
	return &APIClient{
		baseURL: c.DaytonaApiUrl,
		token:   c.ApiToken,
		httpClient: &http.Client{
			Transport: otelhttp.NewTransport(http.DefaultTransport),
		},
	}, nil
}

func (c *APIClient) Do(ctx context.Context, method, path string, body proto.Message, result proto.Message) (*http.Response, error) {
	var bodyReader io.Reader
	if body != nil {
		data, err := protojson.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("marshal request: %w", err)
		}
		bodyReader = bytes.NewReader(data)
	}

	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, bodyReader)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("X-Daytona-Source", "runner")
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return resp, err
	}

	respBody, err := io.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		return resp, fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode >= 400 {
		return resp, fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(respBody))
	}

	if result != nil && len(respBody) > 0 {
		opts := protojson.UnmarshalOptions{DiscardUnknown: true, AllowPartial: true}
		if err := opts.Unmarshal(respBody, result); err != nil {
			return resp, fmt.Errorf("unmarshal response: %w", err)
		}
	}

	return resp, nil
}
