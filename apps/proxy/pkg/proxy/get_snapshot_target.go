// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package proxy

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"regexp"
	"time"

	"github.com/daytonaio/apiclient"
	common_errors "github.com/daytonaio/common-go/pkg/errors"
	"github.com/gin-gonic/gin"

	log "github.com/sirupsen/logrus"
)

func (p *Proxy) getSnapshotTarget(ctx *gin.Context) (*url.URL, map[string]string, error) {
	// Extract snapshot ID from the path
	match := regexp.MustCompile(`^/snapshots/([\w-]+)/build-logs$`).FindStringSubmatch(ctx.Request.URL.Path)
	if len(match) != 2 {
		ctx.Error(common_errors.NewBadRequestError(errors.New("snapshot ID is required")))
		return nil, nil, errors.New("snapshot ID is required")
	}

	snapshotId := match[1]

	snapshot, err := p.getSnapshot(ctx, snapshotId)
	if err != nil {
		ctx.Error(common_errors.NewBadRequestError(fmt.Errorf("failed to get snapshot: %w", err)))
		return nil, nil, fmt.Errorf("failed to get snapshot: %w", err)
	}

	if snapshot.Ref == nil {
		ctx.Error(common_errors.NewBadRequestError(errors.New("snapshot has no snapshot reference")))
		return nil, nil, errors.New("snapshot has no snapshot reference")
	}

	if snapshot.InitialRunnerId == nil {
		ctx.Error(common_errors.NewBadRequestError(errors.New("snapshot has no initial runner")))
		return nil, nil, errors.New("snapshot has no initial runner")
	}

	runnerInfo, err := p.getRunnerInfo(ctx, *snapshot.InitialRunnerId)
	if err != nil {
		ctx.Error(common_errors.NewBadRequestError(fmt.Errorf("failed to get runner info: %w", err)))
		return nil, nil, fmt.Errorf("failed to get runner info: %w", err)
	}

	queryParams := ctx.Request.URL.Query()
	queryParams.Add("snapshotRef", *snapshot.Ref)

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

func (p *Proxy) getSnapshot(ctx context.Context, snapshotId string) (*apiclient.SnapshotDto, error) {
	snapshot, _, err := p.apiclient.SnapshotsAPI.GetSnapshot(ctx, snapshotId).Execute()
	if err != nil {
		return nil, err
	}

	return snapshot, nil
}

func (p *Proxy) getRunnerInfo(ctx context.Context, runnerId string) (*RunnerInfo, error) {
	has, err := p.runnerCache.Has(ctx, runnerId)
	if err != nil {
		return nil, err
	}

	if has {
		return p.runnerCache.Get(ctx, runnerId)
	}

	runner, _, err := p.apiclient.RunnersAPI.GetRunnerFullById(ctx, runnerId).Execute()
	if err != nil {
		return nil, err
	}

	if runner.ApiUrl == nil {
		return nil, errors.New("runner API URL not found")
	}

	info := RunnerInfo{
		ApiUrl: *runner.ApiUrl,
		ApiKey: runner.ApiKey,
	}

	err = p.runnerCache.Set(ctx, runnerId, info, 2*time.Minute)
	if err != nil {
		log.Errorf("Failed to set runner info in cache: %v", err)
	}

	return &info, nil
}
