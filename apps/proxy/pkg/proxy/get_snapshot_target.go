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

	common_errors "github.com/daytonaio/common-go/pkg/errors"
	"github.com/daytonaio/common-go/pkg/utils"
	apiclient "github.com/daytonaio/daytona/libs/api-client-go"
	"github.com/gin-gonic/gin"

	log "github.com/sirupsen/logrus"
)

func (p *Proxy) getSnapshotTarget(ctx *gin.Context) (*url.URL, map[string]string, error) {
	// Extract snapshot ID from the path
	match := regexp.MustCompile(`^/snapshots/([\w-]+)/build-logs$`).FindStringSubmatch(ctx.Request.URL.Path)
	if len(match) != 2 {
		err := common_errors.NewBadRequestError(errors.New("snapshot ID is required"))
		ctx.Error(err)
		return nil, nil, err
	}

	snapshotId := match[1]

	snapshot, err := p.getSnapshot(ctx, snapshotId)
	if err != nil {
		err := classifyUpstreamError(fmt.Errorf("failed to get snapshot: %w", err))
		ctx.Error(err)
		return nil, nil, err
	}

	if snapshot.Ref == nil {
		err := common_errors.NewBadRequestError(errors.New("snapshot has no snapshot reference"))
		ctx.Error(err)
		return nil, nil, err
	}

	if snapshot.InitialRunnerId == nil {
		err := common_errors.NewBadRequestError(errors.New("snapshot has no initial runner"))
		ctx.Error(err)
		return nil, nil, err
	}

	runnerInfo, err := p.getRunnerInfo(ctx, *snapshot.InitialRunnerId)
	if err != nil {
		err := classifyUpstreamError(fmt.Errorf("failed to get runner info: %w", err))
		ctx.Error(err)
		return nil, nil, err
	}

	queryParams := ctx.Request.URL.Query()
	queryParams.Add("snapshotRef", *snapshot.Ref)

	// Build the target URL
	targetURL := fmt.Sprintf("%s/snapshots/logs", runnerInfo.ApiUrl)

	// Create the complete target URL with path
	target, err := url.Parse(targetURL)
	if err != nil {
		err := common_errors.NewInternalServerError(fmt.Errorf("failed to parse target URL: %w", err))
		ctx.Error(err)
		return nil, nil, err
	}
	target.RawQuery = queryParams.Encode()

	return target, map[string]string{
		"X-Daytona-Authorization": fmt.Sprintf("Bearer %s", runnerInfo.ApiKey),
		"X-Forwarded-Host":        ctx.Request.Host,
	}, nil
}

func (p *Proxy) getSnapshot(ctx *gin.Context, snapshotId string) (*apiclient.SnapshotDto, error) {
	var snapshot *apiclient.SnapshotDto
	bearerToken := p.getBearerToken(ctx)
	apiClient := p.getUserApiClient(ctx, bearerToken)

	err := utils.RetryWithExponentialBackoff(ctx, "getSnapshot", proxyMaxRetries, proxyBaseDelay, proxyMaxDelay, func() error {
		s, _, e := apiClient.SnapshotsAPI.GetSnapshot(ctx, snapshotId).Execute()
		snapshot = s
		openapiErr := common_errors.ConvertOpenAPIError(e)

		if openapiErr != nil && !common_errors.IsRetryableOpenAPIError(openapiErr) {
			return &utils.NonRetryableError{Err: openapiErr}
		}

		return openapiErr
	})
	return snapshot, err
}

func (p *Proxy) getRunnerInfo(ctx context.Context, runnerId string) (*RunnerInfo, error) {
	runnerInfo, err := p.runnerCache.Get(ctx, runnerId)
	if err == nil {
		return runnerInfo, nil
	}

	var runner *apiclient.RunnerFull
	err = utils.RetryWithExponentialBackoff(ctx, "getRunnerInfo", proxyMaxRetries, proxyBaseDelay, proxyMaxDelay, func() error {
		r, _, e := p.apiclient.RunnersAPI.GetRunnerFullById(context.Background(), runnerId).Execute()
		runner = r
		openapiErr := common_errors.ConvertOpenAPIError(e)

		if openapiErr != nil && !common_errors.IsRetryableOpenAPIError(openapiErr) {
			return &utils.NonRetryableError{Err: openapiErr}
		}

		return openapiErr
	})
	if err != nil {
		return nil, err
	}

	if runner.ApiUrl == nil {
		return nil, NewRunnerUnreachableError("runner API URL not found")
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
