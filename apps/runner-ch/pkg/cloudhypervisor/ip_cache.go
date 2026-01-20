// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package cloudhypervisor

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
)

// IPCache stores sandbox IP addresses in memory for fast lookups
type IPCache struct {
	mu    sync.RWMutex
	cache map[string]ipEntry
}

type ipEntry struct {
	ip        string
	timestamp time.Time
}

var (
	globalIPCache *IPCache
	ipCacheOnce   sync.Once
)

// GetIPCache returns the global IP cache instance
func GetIPCache() *IPCache {
	ipCacheOnce.Do(func() {
		globalIPCache = &IPCache{
			cache: make(map[string]ipEntry),
		}
		log.Info("IP cache initialized")
	})
	return globalIPCache
}

// Set stores an IP address for a sandbox
func (c *IPCache) Set(sandboxId string, ip string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.cache[sandboxId] = ipEntry{
		ip:        ip,
		timestamp: time.Now(),
	}
	log.Infof("IP cache: stored %s -> %s", sandboxId, ip)
}

// Get retrieves the cached IP for a sandbox
// Returns empty string if not found
func (c *IPCache) Get(sandboxId string) string {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if entry, exists := c.cache[sandboxId]; exists {
		return entry.ip
	}
	return ""
}

// Delete removes the IP entry for a sandbox
func (c *IPCache) Delete(sandboxId string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if _, exists := c.cache[sandboxId]; exists {
		delete(c.cache, sandboxId)
		log.Infof("IP cache: removed %s", sandboxId)
	}
}

// GetOrFetch returns the cached IP, or fetches it from IP pool/sandbox info
func (c *IPCache) GetOrFetch(ctx context.Context, sandboxId string, client *Client) string {
	// Check cache first
	if ip := c.Get(sandboxId); ip != "" {
		return ip
	}

	// Try IP pool (instant lookup)
	if ip := client.GetIPPool().Get(sandboxId); ip != "" {
		c.Set(sandboxId, ip)
		return ip
	}

	// Try to read from stored file (for legacy sandboxes)
	ipFilePath := filepath.Join(client.config.SandboxesPath, sandboxId, "ip")
	if output, err := client.runSSHCommand(ctx, fmt.Sprintf("cat %s 2>/dev/null", ipFilePath)); err == nil {
		if ip := strings.TrimSpace(output); isValidIP(ip) {
			c.Set(sandboxId, ip)
			return ip
		}
	}

	log.Warnf("IP cache: could not find IP for sandbox %s", sandboxId)
	return ""
}

// Clear removes all cached IPs
func (c *IPCache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.cache = make(map[string]ipEntry)
	log.Info("IP cache: cleared all entries")
}

// Size returns the number of cached entries
func (c *IPCache) Size() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.cache)
}

// isValidIP checks if the string looks like a valid IPv4 address
func isValidIP(s string) bool {
	if s == "" {
		return false
	}
	parts := strings.Split(s, ".")
	if len(parts) != 4 {
		return false
	}
	for _, part := range parts {
		// Check it's a number 0-255
		if len(part) == 0 || len(part) > 3 {
			return false
		}
		for _, c := range part {
			if c < '0' || c > '9' {
				return false
			}
		}
	}
	return true
}
