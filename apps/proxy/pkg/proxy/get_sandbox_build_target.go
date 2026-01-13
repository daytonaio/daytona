// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package proxy

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"regexp"

	"github.com/daytonaio/apiclient"
	common_errors "github.com/daytonaio/common-go/pkg/errors"
	"github.com/gin-gonic/gin"
)

func (p *Proxy) getSandboxBuildTarget(ctx *gin.Context) (*url.URL, map[string]string, error) {
	// Extract sandbox ID from the path
	match := regexp.MustCompile(`^/sandboxes/([\w-]+)/build-logs$`).FindStringSubmatch(ctx.Request.URL.Path)
	if len(match) != 2 {
		ctx.Error(common_errors.NewBadRequestError(errors.New("sandbox ID is required")))
		return nil, nil, errors.New("sandbox ID is required")
	}

	sandboxId := match[1]

	sandbox, err := p.getSandbox(ctx, sandboxId)
	if err != nil {
		ctx.Error(common_errors.NewBadRequestError(fmt.Errorf("failed to get sandbox: %w", err)))
		return nil, nil, fmt.Errorf("failed to get sandbox: %w", err)
	}

	if sandbox.BuildInfo == nil {
		ctx.Error(common_errors.NewBadRequestError(errors.New("sandbox has no build info")))
		return nil, nil, errors.New("sandbox has no build info")
	}

	runnerInfo, err := p.getRunnerInfo(ctx, *sandbox.RunnerId)
	if err != nil {
		ctx.Error(common_errors.NewBadRequestError(fmt.Errorf("failed to get runner info: %w", err)))
		return nil, nil, fmt.Errorf("failed to get runner info: %w", err)
	}

	queryParams := ctx.Request.URL.Query()
	queryParams.Add("snapshotRef", sandbox.BuildInfo.SnapshotRef)

	// Build the target URL
	targetURL := fmt.Sprintf("%s/snapshots/logs", runnerInfo.ApiUrl)

	// Create the complete target URL with path
	target, err := url.Parse(targetURL)
	if err != nil {
		ctx.Error(common_errors.NewBadRequestError(fmt.Errorf("failed to parse target URL: %w", err)))
		return nil, nil, fmt.Errorf("failed to parse target URL: %w", err)
	}
	target.RawQuery = queryParams.Encode()

	return target, map[string]string{
		"X-Daytona-Authorization": fmt.Sprintf("Bearer %s", runnerInfo.ApiKey),
		"X-Forwarded-Host":        ctx.Request.Host,
	}, nil
}

func (p *Proxy) getSandbox(ctx context.Context, sandboxId string) (*apiclient.Sandbox, error) {
	sandbox, _, err := p.apiclient.SandboxAPI.GetSandbox(ctx, sandboxId).Execute()
	if err != nil {
		return nil, err
	}

	return sandbox, nil
}
