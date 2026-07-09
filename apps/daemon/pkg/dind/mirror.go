// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

// Package dind configures the nested Docker daemon that users run inside a
// sandbox (Docker-in-Docker). Its main job is to point that daemon at a
// registry pull-through cache so that `docker pull` of public images does not
// hit Docker Hub's anonymous pull rate limit (HTTP 429). Because many sandboxes
// on a single runner share one outbound NAT IP, the shared anonymous quota is
// exhausted quickly; a pull-through cache (optionally authenticated to Docker
// Hub) both raises that quota and serves repeated pulls straight from cache.
package dind

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/url"
	"os"
	"path/filepath"
	"sort"
)

// DaemonConfigPath is where the Docker daemon reads its configuration from.
// dockerd loads this file on startup, so writing it before the user starts
// their nested daemon is enough for the mirror to take effect.
const DaemonConfigPath = "/etc/docker/daemon.json"

// ConfigureRegistryMirror writes mirrorURL into the nested Docker daemon's
// configuration file so that Docker-in-Docker pulls are served through the
// pull-through cache.
//
// It is intentionally best-effort and conservative:
//   - When mirrorURL is empty the feature is disabled and it is a no-op.
//   - An existing user/base-image "registry-mirrors" entry is never overwritten,
//     so callers who deliberately configured their own mirror keep it.
//   - Any existing daemon.json content is preserved (only the mirror-related
//     keys are added).
//
// It returns an error describing what went wrong, but the caller should treat
// failures as non-fatal: a sandbox may not run Docker at all, or the daemon may
// lack permission to write /etc/docker.
func ConfigureRegistryMirror(logger *slog.Logger, mirrorURL string) error {
	return configureRegistryMirrorAt(logger, DaemonConfigPath, mirrorURL)
}

// configureRegistryMirrorAt is the path-parameterized implementation behind
// ConfigureRegistryMirror, split out so tests can target a temporary file.
func configureRegistryMirrorAt(logger *slog.Logger, path, mirrorURL string) error {
	if mirrorURL == "" {
		return nil
	}

	if _, err := url.ParseRequestURI(mirrorURL); err != nil {
		return fmt.Errorf("invalid registry mirror URL %q: %w", mirrorURL, err)
	}

	existing, err := os.ReadFile(path)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to read %s: %w", path, err)
	}

	updated, changed, err := mergeRegistryMirror(existing, mirrorURL)
	if err != nil {
		return err
	}

	if !changed {
		if logger != nil {
			logger.Debug("Docker daemon already configures a registry mirror; leaving it unchanged", "path", path)
		}
		return nil
	}

	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return fmt.Errorf("failed to create %s: %w", filepath.Dir(path), err)
	}

	if err := os.WriteFile(path, updated, 0644); err != nil {
		return fmt.Errorf("failed to write %s: %w", path, err)
	}

	if logger != nil {
		logger.Info("Configured Docker-in-Docker registry mirror", "path", path, "mirror", mirrorURL)
	}

	return nil
}

// mergeRegistryMirror merges a registry mirror into the raw daemon.json bytes.
// It returns the new content, whether anything changed, and an error. It is
// pure (no filesystem access) to keep it easy to unit test.
//
// The merge rules are:
//   - If the existing config already sets a non-empty "registry-mirrors", the
//     config is returned unchanged (changed=false).
//   - Otherwise the mirror is added as the sole "registry-mirrors" entry.
//   - When the mirror is served over plain HTTP, its host is also added to
//     "insecure-registries" (dockerd otherwise refuses to use an HTTP mirror),
//     without dropping any hosts the user already listed there.
func mergeRegistryMirror(existing []byte, mirrorURL string) ([]byte, bool, error) {
	parsed, err := url.ParseRequestURI(mirrorURL)
	if err != nil {
		return nil, false, fmt.Errorf("invalid registry mirror URL %q: %w", mirrorURL, err)
	}

	config := map[string]any{}
	if len(existing) > 0 {
		if err := json.Unmarshal(existing, &config); err != nil {
			return nil, false, fmt.Errorf("failed to parse existing %s: %w", DaemonConfigPath, err)
		}
	}

	if mirrors, ok := config["registry-mirrors"].([]any); ok && len(mirrors) > 0 {
		return existing, false, nil
	}

	config["registry-mirrors"] = []string{mirrorURL}

	// dockerd rejects an HTTP mirror unless the host is trusted as insecure.
	if parsed.Scheme == "http" {
		config["insecure-registries"] = addInsecureRegistry(config["insecure-registries"], parsed.Host)
	}

	out, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return nil, false, fmt.Errorf("failed to encode %s: %w", DaemonConfigPath, err)
	}
	out = append(out, '\n')

	return out, true, nil
}

// addInsecureRegistry returns the union of the already-configured insecure
// registries and host, preserving existing entries and avoiding duplicates.
func addInsecureRegistry(current any, host string) []string {
	seen := map[string]struct{}{}
	result := []string{}

	if list, ok := current.([]any); ok {
		for _, entry := range list {
			if s, ok := entry.(string); ok {
				if _, dup := seen[s]; !dup {
					seen[s] = struct{}{}
					result = append(result, s)
				}
			}
		}
	}

	if _, dup := seen[host]; !dup {
		result = append(result, host)
	}

	sort.Strings(result)
	return result
}
