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

func (p *Proxy) GetProxyTarget(ctx *gin.Context, toolboxSubpathRequest bool) (*url.URL, map[string]string, error) {
	var targetPort, targetPath, sandboxID string

	if toolboxSubpathRequest {
		// Expected format: /toolbox/<sandboxID>/<targetPath>
		var err error
		targetPort, sandboxID, targetPath, err = p.parseToolboxSubpath(ctx.Param("path"))
		if err != nil {
			ctx.Error(common_errors.NewBadRequestError(err))
			return nil, nil, err
		}
	} else {
		// Extract port and sandbox ID from the host header
		// Expected format: 1234-some-id-uuid.proxy.domain
		var err error
		targetPort, sandboxID, err = p.parseHost(ctx.Request.Host)
		if err != nil {
			ctx.Error(common_errors.NewBadRequestError(err))
			return nil, nil, err
		}

		targetPath = ctx.Param("path")
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

	if !*isPublic || targetPort == TERMINAL_PORT || targetPort == TOOLBOX_PORT {
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

	// Skip last activity update if header is set
	if ctx.Request.Header.Get(SKIP_LAST_ACTIVITY_UPDATE_HEADER) != "true" {
		p.updateLastActivity(ctx.Request.Context(), sandboxID, true)
		ctx.Request.Header.Del(SKIP_LAST_ACTIVITY_UPDATE_HEADER)
	}

	// Build the target URL
	targetURL := fmt.Sprintf("%s/sandboxes/%s/toolbox/proxy/%s", runnerInfo.ApiUrl, sandboxID, targetPort)
	if toolboxSubpathRequest {
		targetURL = fmt.Sprintf("%s/sandboxes/%s/toolbox", runnerInfo.ApiUrl, sandboxID)
	}

	// Ensure path always has a leading slash but not duplicate slashes
	if targetPath == "" {
		targetPath = "/"
	} else if !strings.HasPrefix(targetPath, "/") {
		targetPath = "/" + targetPath
	}

	// Create the complete target URL with path
	target, err := url.Parse(fmt.Sprintf("%s%s", targetURL, targetPath))
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
		ApiUrl: runner.ProxyUrl,
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
	apiValidation := func() bool {
		_, resp, _ := p.apiclient.PreviewAPI.IsValidAuthToken(context.Background(), sandboxId, authKey).Execute()
		return resp != nil && resp.StatusCode == http.StatusOK
	}

	return p.validateAndCache(ctx, sandboxId, authKey, apiValidation)
}

func (p *Proxy) getSandboxBearerTokenValid(ctx context.Context, sandboxId string, bearerToken string) (*bool, error) {
	apiValidation := func() bool {
		return p.hasSandboxAccess(ctx, sandboxId, bearerToken)
	}

	return p.validateAndCache(ctx, sandboxId, bearerToken, apiValidation)
}

func (p *Proxy) validateAndCache(
	ctx context.Context,
	sandboxId string,
	authKey string,
	apiValidation func() bool,
) (*bool, error) {
	cacheKey := fmt.Sprintf("%s:%s", sandboxId, authKey)
	has, err := p.sandboxAuthKeyValidCache.Has(ctx, cacheKey)
	if err != nil {
		return nil, err
	}

	if has {
		return p.sandboxAuthKeyValidCache.Get(ctx, cacheKey)
	}

	isValid := apiValidation()

	if err := p.sandboxAuthKeyValidCache.Set(ctx, cacheKey, isValid, 2*time.Minute); err != nil {
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

func (p *Proxy) updateLastActivity(ctx context.Context, sandboxId string, shouldPollUpdate bool) {
	// Prevent frequent updates by caching the last update
	cached, err := p.sandboxLastActivityUpdateCache.Has(ctx, sandboxId)
	if err != nil {
		// If cache doesn't work, skip the update to avoid spamming the API
		log.Errorf("failed to check last activity update cache for sandbox %s: %v", sandboxId, err)
		return
	}

	if !cached {
		_, err := p.apiclient.SandboxAPI.UpdateLastActivity(ctx, sandboxId).Execute()
		if err != nil {
			log.Errorf("failed to update last activity for sandbox %s: %v", sandboxId, err)
			return
		}

		err = p.sandboxLastActivityUpdateCache.Set(ctx, sandboxId, true, 45*time.Second)
		if err != nil {
			log.Errorf("failed to set last activity update cache for sandbox %s: %v", sandboxId, err)
		}
	}

	if shouldPollUpdate {
		// Update keep alive every 45 seconds until the request is done
		go func() {
			ticker := time.NewTicker(45 * time.Second)
			defer ticker.Stop()

			for {
				select {
				case <-ticker.C:
					p.updateLastActivity(ctx, sandboxId, false)
				case <-ctx.Done():
					return
				}
			}
		}()
	}
}

func (p *Proxy) parseToolboxSubpath(path string) (string, string, string, error) {
	// Expected format: /toolbox/<sandboxID>/<path>
	if path == "" {
		return "", "", "", errors.New("path is required")
	}

	if !strings.HasPrefix(path, "/toolbox/") {
		return "", "", "", errors.New("path must start with /toolbox/")
	}

	// Trim prefix and split by "/"
	parts := strings.SplitN(strings.TrimPrefix(path, "/toolbox/"), "/", 2)
	if len(parts) < 2 {
		return "", "", "", errors.New("path must be of format /toolbox/<sandboxId>/<path>")
	}

	sandboxID := parts[0]
	targetPath := "/" + parts[1]

	return TOOLBOX_PORT, sandboxID, targetPath, nil
}
