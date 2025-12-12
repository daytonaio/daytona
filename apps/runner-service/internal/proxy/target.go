// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package proxy

import (
	"context"
	"errors"
	"fmt"
	"net/url"
)

func (s *ProxyServer) getProxyTarget(ctx context.Context, sandboxId string) (*url.URL, error) {
	cacheItem := s.cache.Get(sandboxId)
	sandboxIp := ""

	if cacheItem != nil {
		sandboxIp = cacheItem.Value().IP
		s.cache.Set(sandboxId, &ProxyCacheItem{IP: sandboxIp}, s.cacheTTL)
	} else {
		container, err := s.dockerClient.ContainerInspect(ctx, sandboxId)
		if err != nil {
			s.log.ErrorContext(ctx, "Error getting sandbox from docker client", "error", err)
			return nil, fmt.Errorf("sandbox not found: %w", err)
		}

		network, ok := container.NetworkSettings.Networks[s.network]
		if !ok {
			return nil, fmt.Errorf("sandbox not connected to network %q", s.network)
		}

		sandboxIp = network.IPAddress
		if sandboxIp == "" {
			return nil, errors.New("no IP address found. Is the Sandbox started?")
		}

		s.cache.Set(sandboxId, &ProxyCacheItem{IP: sandboxIp}, s.cacheTTL)
	}

	targetURL := fmt.Sprintf("http://%s:%d", sandboxIp, s.targetPort)
	target, err := url.Parse(targetURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse target URL: %w", err)
	}

	return target, nil
}
