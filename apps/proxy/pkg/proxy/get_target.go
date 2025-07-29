// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package proxy

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	common_errors "github.com/daytonaio/common-go/pkg/errors"
	"github.com/gin-gonic/gin"

	log "github.com/sirupsen/logrus"
)

func (p *Proxy) GetProxyTarget(ctx *gin.Context) (*url.URL, map[string]string, error) {
	// Extract port and sandbox ID from the host header
	// Expected format: 1234-some-id-uuid.proxy.domain
	targetPort, sandboxID, err := p.parseHost(ctx.Request.Host)
	if err != nil {
		ctx.Error(common_errors.NewBadRequestError(err))
		return nil, nil, err
	}

	if targetPort == "" {
		ctx.Error(common_errors.NewBadRequestError(errors.New("target port is required")))
		return nil, nil, errors.New("target port is required")
	}

	if sandboxID == "" {
		ctx.Error(common_errors.NewBadRequestError(errors.New("sandbox ID is required")))
		return nil, nil, errors.New("sandbox ID is required")
	}

	isPublic, err := p.getSandboxPublic(ctx, sandboxID)
	if err != nil {
		ctx.Error(common_errors.NewBadRequestError(fmt.Errorf("failed to get sandbox public status: %w", err)))
		return nil, nil, fmt.Errorf("failed to get sandbox public status: %w", err)
	}

	if !*isPublic || targetPort == "22222" {
		err, didRedirect := p.Authenticate(ctx, sandboxID)
		if err != nil {
			if !didRedirect {
				ctx.Error(common_errors.NewUnauthorizedError(err))
			}
			return nil, nil, err
		}
	}

	runnerInfo, err := p.getRunnerInfo(ctx, sandboxID)
	if err != nil {
		ctx.Error(common_errors.NewBadRequestError(fmt.Errorf("failed to get runner info: %w", err)))
		return nil, nil, fmt.Errorf("failed to get runner info: %w", err)
	}

	// Build the target URL
	targetURL := fmt.Sprintf("%s/sandboxes/%s/toolbox/proxy/%s", runnerInfo.ApiUrl, sandboxID, targetPort)

	// Get the wildcard path and normalize it
	path := ctx.Param("path")

	// Ensure path always has a leading slash but not duplicate slashes
	if path == "" {
		path = "/"
	} else if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}

	// Create the complete target URL with path
	target, err := url.Parse(fmt.Sprintf("%s%s", targetURL, path))
	if err != nil {
		ctx.Error(common_errors.NewBadRequestError(fmt.Errorf("failed to parse target URL: %w", err)))
		return nil, nil, fmt.Errorf("failed to parse target URL: %w", err)
	}

	return target, map[string]string{
		"X-Daytona-Authorization": fmt.Sprintf("Bearer %s", runnerInfo.ApiKey),
		"X-Forwarded-Host":        ctx.Request.Host,
	}, nil
}

func (p *Proxy) getRunnerInfo(ctx context.Context, sandboxId string) (*RunnerInfo, error) {
	has, err := p.runnerCache.Has(ctx, sandboxId)
	if err != nil {
		return nil, err
	}

	if has {
		return p.runnerCache.Get(ctx, sandboxId)
	}

	runner, _, err := p.apiclient.RunnersAPI.GetRunnerBySandboxId(context.Background(), sandboxId).Execute()
	if err != nil {
		return nil, err
	}

	info := RunnerInfo{
		ApiUrl: runner.ApiUrl,
		ApiKey: runner.ApiKey,
	}

	err = p.runnerCache.Set(ctx, sandboxId, info, 2*time.Minute)
	if err != nil {
		log.Errorf("Failed to set runner info in cache: %v", err)
	}

	return &info, nil
}

func (p *Proxy) getSandboxPublic(ctx context.Context, sandboxId string) (*bool, error) {
	has, err := p.sandboxPublicCache.Has(ctx, sandboxId)
	if err != nil {
		return nil, err
	}

	if has {
		return p.sandboxPublicCache.Get(ctx, sandboxId)
	}

	isPublic := false
	_, resp, _ := p.apiclient.PreviewAPI.IsSandboxPublic(context.Background(), sandboxId).Execute()
	if resp != nil && resp.StatusCode == http.StatusOK {
		isPublic = true
	}

	err = p.sandboxPublicCache.Set(ctx, sandboxId, isPublic, 1*time.Hour)
	if err != nil {
		log.Errorf("Failed to set sandbox public in cache: %v", err)
	}

	return &isPublic, nil
}

func (p *Proxy) getSandboxAuthKeyValid(ctx context.Context, sandboxId string, authKey string) (*bool, error) {
	has, err := p.sandboxAuthKeyValidCache.Has(ctx, authKey)
	if err != nil {
		return nil, err
	}

	if has {
		return p.sandboxAuthKeyValidCache.Get(ctx, authKey)
	}

	isValid := false
	_, resp, _ := p.apiclient.PreviewAPI.IsValidAuthToken(context.Background(), sandboxId, authKey).Execute()
	if resp != nil && resp.StatusCode == http.StatusOK {
		isValid = true
	}

	err = p.sandboxAuthKeyValidCache.Set(ctx, authKey, isValid, 2*time.Minute)
	if err != nil {
		log.Errorf("Failed to set sandbox auth key valid in cache: %v", err)
	}

	return &isValid, nil
}

func (p *Proxy) parseHost(host string) (targetPort string, sandboxID string, err error) {
	// Extract port and sandbox ID from the host header
	// Expected format: 1234-some-id-uuid.proxy.domain
	if host == "" {
		return "", "", errors.New("host is required")
	}

	// Split the host to extract the port and sandbox ID
	parts := strings.Split(host, ".")
	if len(parts) == 0 {
		return "", "", errors.New("invalid host format")
	}

	// Extract port from the first part (e.g., "1234-some-id-uuid")
	hostPrefix := parts[0]
	dashIndex := strings.Index(hostPrefix, "-")
	if dashIndex == -1 {
		return "", "", errors.New("invalid host format: port and sandbox ID not found")
	}

	targetPort = hostPrefix[:dashIndex]
	sandboxID = hostPrefix[dashIndex+1:]

	return targetPort, sandboxID, nil
}
