// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package proxy

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	common_errors "github.com/daytonaio/common-go/pkg/errors"
	"github.com/daytonaio/common-go/pkg/utils"
	apiclient "github.com/daytonaio/daytona/libs/api-client-go"
	"github.com/gin-gonic/gin"

	log "github.com/sirupsen/logrus"
)

func (p *Proxy) GetProxyTarget(ctx *gin.Context) (*url.URL, map[string]string, error) {
	var targetPort, targetPath, sandboxIdOrSignedToken string

	if ctx.GetBool(IS_TOOLBOX_REQUEST_KEY) {
		// Expected format: /toolbox/<sandboxID>/<targetPath>
		var err error
		targetPort, sandboxIdOrSignedToken, targetPath, err = p.parseToolboxSubpath(ctx.Param("path"))
		if err != nil {
			ctx.Error(common_errors.NewBadRequestError(err))
			return nil, nil, err
		}
	} else {
		// Extract port and sandbox ID from the host header
		// Expected format: 1234-<sandboxId | token>.proxy.domain
		var err error
		targetPort, sandboxIdOrSignedToken, _, err = p.parseHost(ctx.Request.Host)
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

	if sandboxIdOrSignedToken == "" {
		ctx.Error(common_errors.NewBadRequestError(errors.New("sandbox ID or signed token is required")))
		return nil, nil, errors.New("sandbox ID or signed token is required")
	}

	sandboxId := sandboxIdOrSignedToken

	isPublic, err := p.getSandboxPublic(ctx, sandboxIdOrSignedToken)
	if err != nil {
		ctx.Error(common_errors.NewBadRequestError(fmt.Errorf("failed to get sandbox public status: %w", err)))
		return nil, nil, fmt.Errorf("failed to get sandbox public status: %w", err)
	}

	if !*isPublic || targetPort == TERMINAL_PORT || targetPort == TOOLBOX_PORT || targetPort == RECORDING_DASHBOARD_PORT {
		portFloat, err := strconv.ParseFloat(targetPort, 64)
		if err != nil {
			ctx.Error(common_errors.NewBadRequestError(fmt.Errorf("failed to parse target port: %w", err)))
			return nil, nil, fmt.Errorf("failed to parse target port: %w", err)
		}
		var didRedirect bool
		sandboxId, didRedirect, err = p.Authenticate(ctx, sandboxIdOrSignedToken, float32(portFloat))
		if err != nil {
			if !didRedirect {
				ctx.Error(err)
			}
			return nil, nil, err
		}
	}

	runnerInfo, err := p.getSandboxRunnerInfo(ctx, sandboxId)
	if err != nil {
		ctx.Error(common_errors.NewBadRequestError(fmt.Errorf("failed to get runner info: %w", err)))
		return nil, nil, fmt.Errorf("failed to get runner info: %w", err)
	}

	// Skip last activity update if header is set
	if ctx.Request.Header.Get(SKIP_LAST_ACTIVITY_UPDATE_HEADER) != "true" {
		doneCh := make(chan struct{})
		go p.updateLastActivity(ctx.Request.Context(), sandboxId, true, doneCh)
		ctx.Request.Header.Del(SKIP_LAST_ACTIVITY_UPDATE_HEADER)
		ctx.Set(ACTIVITY_POLL_STOP_KEY, func() {
			close(doneCh)
		})
	}

	// Build the target URL
	targetURL := fmt.Sprintf("%s/sandboxes/%s/toolbox/proxy/%s", runnerInfo.ApiUrl, sandboxId, targetPort)
	if ctx.GetBool(IS_TOOLBOX_REQUEST_KEY) {
		targetURL = fmt.Sprintf("%s/sandboxes/%s/toolbox", runnerInfo.ApiUrl, sandboxId)
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

func (p *Proxy) getSandboxRunnerInfo(ctx context.Context, sandboxId string) (*RunnerInfo, error) {
	runnerInfo, err := p.sandboxRunnerCache.Get(ctx, sandboxId)
	if err == nil {
		return runnerInfo, nil
	}

	var runner *apiclient.RunnerFull
	err = utils.RetryWithExponentialBackoff(ctx, "getSandboxRunnerInfo", proxyMaxRetries, proxyBaseDelay, proxyMaxDelay, func() error {
		r, _, e := p.apiclient.RunnersAPI.GetRunnerBySandboxId(context.Background(), sandboxId).Execute()
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

	if runner.ProxyUrl == nil {
		return nil, errors.New("runner proxy URL not found")
	}

	info := RunnerInfo{
		ApiUrl: *runner.ProxyUrl,
		ApiKey: runner.ApiKey,
	}

	err = p.sandboxRunnerCache.Set(ctx, sandboxId, info, 2*time.Minute)
	if err != nil {
		log.Errorf("Failed to set runner info in cache: %v", err)
	}

	return &info, nil
}

func (p *Proxy) getSandboxPublic(ctx context.Context, sandboxId string) (*bool, error) {
	isPublicCache, err := p.sandboxPublicCache.Get(ctx, sandboxId)
	if err == nil {
		return isPublicCache, nil
	}

	var isPublic bool
	err = utils.RetryWithExponentialBackoff(ctx, "getSandboxPublic", proxyMaxRetries, proxyBaseDelay, proxyMaxDelay, func() error {
		_, res, err := p.apiclient.PreviewAPI.IsSandboxPublic(context.Background(), sandboxId).Execute()
		if res != nil && res.StatusCode == http.StatusOK {
			isPublic = true
			return nil
		}
		openapiErr := common_errors.ConvertOpenAPIError(err)

		if openapiErr != nil {
			if res != nil && res.StatusCode >= 400 && res.StatusCode < 500 &&
				res.StatusCode != http.StatusRequestTimeout && res.StatusCode != http.StatusTooManyRequests {
				isPublic = false
				return nil
			}
			if !common_errors.IsRetryableOpenAPIError(openapiErr) {
				return &utils.NonRetryableError{Err: openapiErr}
			}
			return openapiErr
		}
		isPublic = false
		return nil
	})
	if err != nil {
		return nil, err
	}

	if cacheErr := p.sandboxPublicCache.Set(ctx, sandboxId, isPublic, 1*time.Hour); cacheErr != nil {
		log.Errorf("Failed to set sandbox public in cache: %v", cacheErr)
	}

	return &isPublic, nil
}

func (p *Proxy) getSandboxAuthKeyValid(ctx context.Context, sandboxId string, authKey string) (*bool, error) {
	apiValidation := func() (bool, error) {
		_, resp, err := p.apiclient.PreviewAPI.IsValidAuthToken(context.Background(), sandboxId, authKey).Execute()
		if resp != nil && resp.StatusCode == http.StatusOK {
			return true, nil
		}
		openapiErr := common_errors.ConvertOpenAPIError(err)

		if openapiErr != nil {
			if resp != nil && resp.StatusCode >= 400 && resp.StatusCode < 500 &&
				resp.StatusCode != http.StatusRequestTimeout && resp.StatusCode != http.StatusTooManyRequests {
				return false, nil
			}
			if !common_errors.IsRetryableOpenAPIError(openapiErr) {
				return false, &utils.NonRetryableError{Err: openapiErr}
			}
			return false, openapiErr
		}
		return false, nil
	}

	return p.validateAndCache(ctx, sandboxId, authKey, apiValidation)
}

func (p *Proxy) getSandboxBearerTokenValid(ctx context.Context, sandboxId string, bearerToken string) (*bool, error) {
	apiValidation := func() (bool, error) {
		return p.hasSandboxAccess(ctx, sandboxId, bearerToken)
	}

	return p.validateAndCache(ctx, sandboxId, bearerToken, apiValidation)
}

func (p *Proxy) validateAndCache(
	ctx context.Context,
	sandboxId string,
	authKey string,
	apiValidation func() (bool, error),
) (*bool, error) {
	cacheKey := fmt.Sprintf("%s:%s", sandboxId, authKey)
	authKeyValidCache, err := p.sandboxAuthKeyValidCache.Get(ctx, cacheKey)
	if err == nil {
		return authKeyValidCache, nil
	}

	var isValid bool
	validationErr := utils.RetryWithExponentialBackoff(ctx, "validateAndCache", proxyMaxRetries, proxyBaseDelay, proxyMaxDelay, func() error {
		result, err := apiValidation()
		if err != nil {
			return err
		}
		isValid = result
		return nil
	})
	if validationErr != nil {
		return nil, validationErr
	}

	cacheTTL := 2 * time.Minute
	if !isValid {
		cacheTTL = 5 * time.Second
	}
	if err := p.sandboxAuthKeyValidCache.Set(ctx, cacheKey, isValid, cacheTTL); err != nil {
		log.Errorf("Failed to set sandbox auth key valid in cache: %v", err)
	}

	return &isValid, nil
}

func (p *Proxy) parseHost(host string) (targetPort string, sandboxIdOrSignedToken string, baseHost string, err error) {
	// Extract port and sandbox ID from the host header
	// Expected format: 1234-some-id-uuid.proxy.domain
	if host == "" {
		return "", "", "", errors.New("host is required")
	}

	// Split the host to extract the port and sandbox ID
	parts := strings.Split(host, ".")
	if len(parts) == 0 {
		return "", "", "", errors.New("invalid host format")
	}

	if len(parts) < 2 {
		return "", "", "", errors.New("invalid host format: must have subdomain")
	}

	// Extract port from the first part (e.g., "1234-some-id-uuid")
	hostPrefix := parts[0]
	before, after, ok := strings.Cut(hostPrefix, "-")
	if !ok {
		return "", "", "", errors.New("invalid host format: port and sandbox ID not found")
	}

	targetPort = before

	// Check that port is numeric
	if _, err := strconv.Atoi(targetPort); err != nil {
		return "", "", "", fmt.Errorf("invalid port '%s': must be numeric", targetPort)
	}

	sandboxIdOrSignedToken = after
	// Join remaining parts to form the base domain (e.g., "proxy.domain")
	baseHost = strings.Join(parts[1:], ".")

	return targetPort, sandboxIdOrSignedToken, baseHost, nil
}

// updateLastActivity updates the last activity timestamp for a sandbox.
// If shouldPollUpdate is true, it starts a goroutine that updates every 50 seconds.
func (p *Proxy) updateLastActivity(ctx context.Context, sandboxId string, shouldPollUpdate bool, doneCh chan struct{}) {
	pollInterval := 50 * time.Second

	p.doActivityUpdate(ctx, sandboxId, pollInterval)

	if shouldPollUpdate {
		go func() {
			ticker := time.NewTicker(pollInterval)
			defer ticker.Stop()

			for {
				select {
				case <-ticker.C:
					tickCtx, cancel := context.WithTimeout(context.Background(), time.Duration(p.config.ApiClientTimeoutSec)*time.Second)
					done := make(chan struct{})
					go func() {
						defer close(done)
						p.doActivityUpdate(tickCtx, sandboxId, pollInterval)
					}()
					select {
					case <-done:
						cancel()
					case <-doneCh:
						cancel()
						return
					}
				case <-doneCh:
					return
				}
			}
		}()
	}
}

func (p *Proxy) doActivityUpdate(ctx context.Context, sandboxId string, pollInterval time.Duration) {
	cacheCtx := context.WithoutCancel(ctx)

	cached, err := p.sandboxLastActivityUpdateCache.Has(cacheCtx, sandboxId)
	if err != nil {
		log.Errorf("failed to check last activity update cache for sandbox %s: %v", sandboxId, err)
		return
	}

	if cached {
		return
	}

	_, err = p.apiclient.SandboxAPI.UpdateLastActivity(ctx, sandboxId).Execute()
	if err != nil {
		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			return
		}
		log.Errorf("failed to update last activity for sandbox %s: %v", sandboxId, err)
		return
	}

	err = p.sandboxLastActivityUpdateCache.Set(cacheCtx, sandboxId, true, pollInterval-5*time.Second)
	if err != nil {
		log.Errorf("failed to set last activity update cache for sandbox %s: %v", sandboxId, err)
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
