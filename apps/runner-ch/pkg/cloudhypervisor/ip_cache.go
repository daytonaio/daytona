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

// GetOrFetch returns the cached IP, or fetches it from namespace pool/sandbox info
// With network namespaces, this returns the guest IP (192.168.0.2) which is the same
// for all VMs. Use GetRoutableIP for the IP that can be used to reach the VM from host.
func (c *IPCache) GetOrFetch(ctx context.Context, sandboxId string, client *Client) string {
	// Check cache first
	if ip := c.Get(sandboxId); ip != "" {
		return ip
	}

	// Try network namespace pool (new architecture)
	if ns := client.GetNetNSPool().Get(sandboxId); ns != nil {
		c.Set(sandboxId, ns.GuestIP)
		return ns.GuestIP
	}

	// Try IP pool (legacy, instant lookup)
	if ip := client.GetIPPool().Get(sandboxId); ip != "" {
		c.Set(sandboxId, ip)
		return ip
	}

	// Try to read from stored file (for legacy sandboxes)
	ipFilePath := filepath.Join(client.config.SandboxesPath, sandboxId, "ip")
	if output, err := client.runShellScript(ctx, fmt.Sprintf("cat %s 2>/dev/null", ipFilePath)); err == nil {
		if ip := strings.TrimSpace(output); isValidIP(ip) {
			c.Set(sandboxId, ip)
			return ip
		}
	}

	log.Warnf("IP cache: could not find IP for sandbox %s", sandboxId)
	return ""
}

// GetRoutableIP returns the IP that can be used to reach the VM from the host
// With network namespaces, this is the namespace's external IP (10.0.{num}.1)
// which the host can route to, and the namespace NATs to the guest IP (192.168.0.2)
func (c *IPCache) GetRoutableIP(ctx context.Context, sandboxId string, client *Client) string {
	// With network namespaces, the routable IP is the namespace's external IP
	if ns := client.GetNetNSPool().Get(sandboxId); ns != nil {
		// Return the external IP that the host can reach
		// The namespace will NAT this to the guest IP
		return ns.ExternalIP
	}

	// Fallback to the cached/fetched IP (legacy mode or direct access)
	return c.GetOrFetch(ctx, sandboxId, client)
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
