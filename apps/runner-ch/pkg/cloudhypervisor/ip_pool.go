// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package cloudhypervisor

import (
	"context"
	"fmt"
	"path/filepath"
	"strconv"
	"strings"
	"sync"

	log "github.com/sirupsen/logrus"
)

const (
	// IP pool range: 10.0.0.2 - 10.0.0.254
	IPPoolStart   = 2
	IPPoolEnd     = 254
	IPPoolNetwork = "10.0.0"
	IPPoolGateway = "10.0.0.1"
	IPPoolNetmask = "255.255.255.0"
	IPPoolCIDR    = "24"
)

// IPPool manages a pool of static IP addresses for VMs
type IPPool struct {
	mu        sync.Mutex
	allocated map[string]string // sandboxId -> IP
	available []int             // available last octets
	client    *Client
}

// NewIPPool creates a new IP pool
func NewIPPool(client *Client) *IPPool {
	pool := &IPPool{
		allocated: make(map[string]string),
		available: make([]int, 0, IPPoolEnd-IPPoolStart+1),
		client:    client,
	}

	// Initialize available pool
	for i := IPPoolStart; i <= IPPoolEnd; i++ {
		pool.available = append(pool.available, i)
	}

	return pool
}

// Initialize loads existing allocations from sandbox directories
func (p *IPPool) Initialize(ctx context.Context) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	// Get list of existing sandboxes
	sandboxes, err := p.client.List(ctx)
	if err != nil {
		return fmt.Errorf("failed to list sandboxes: %w", err)
	}

	log.Infof("IP pool: loading allocations for %d existing sandboxes", len(sandboxes))

	for _, sandboxId := range sandboxes {
		// Check if sandbox has stored IP
		ipFilePath := filepath.Join(p.client.config.SandboxesPath, sandboxId, "ip")
		output, err := p.client.runSSHCommand(ctx, fmt.Sprintf("cat %s 2>/dev/null", ipFilePath))
		if err != nil {
			continue
		}

		ip := strings.TrimSpace(output)
		if ip == "" {
			continue
		}

		// Parse the last octet
		parts := strings.Split(ip, ".")
		if len(parts) != 4 {
			continue
		}
		lastOctet, err := strconv.Atoi(parts[3])
		if err != nil {
			continue
		}

		// Mark as allocated
		p.allocated[sandboxId] = ip
		p.removeFromAvailable(lastOctet)
		log.Debugf("IP pool: restored allocation %s -> %s", sandboxId, ip)
	}

	log.Infof("IP pool: %d allocated, %d available", len(p.allocated), len(p.available))
	return nil
}

// Allocate assigns an IP to a sandbox
func (p *IPPool) Allocate(sandboxId string) (string, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	// Check if already allocated
	if ip, exists := p.allocated[sandboxId]; exists {
		return ip, nil
	}

	// Get next available IP
	if len(p.available) == 0 {
		return "", fmt.Errorf("IP pool exhausted")
	}

	lastOctet := p.available[0]
	p.available = p.available[1:]

	ip := fmt.Sprintf("%s.%d", IPPoolNetwork, lastOctet)
	p.allocated[sandboxId] = ip

	log.Infof("IP pool: allocated %s -> %s (%d remaining)", sandboxId, ip, len(p.available))
	return ip, nil
}

// Release returns an IP to the pool
func (p *IPPool) Release(sandboxId string) {
	p.mu.Lock()
	defer p.mu.Unlock()

	ip, exists := p.allocated[sandboxId]
	if !exists {
		return
	}

	// Parse last octet and return to pool
	parts := strings.Split(ip, ".")
	if len(parts) == 4 {
		if lastOctet, err := strconv.Atoi(parts[3]); err == nil {
			p.available = append(p.available, lastOctet)
		}
	}

	delete(p.allocated, sandboxId)
	log.Infof("IP pool: released %s <- %s (%d remaining)", sandboxId, ip, len(p.available))
}

// Get returns the allocated IP for a sandbox
func (p *IPPool) Get(sandboxId string) string {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.allocated[sandboxId]
}

// GetGateway returns the gateway IP
func (p *IPPool) GetGateway() string {
	return IPPoolGateway
}

// GetNetmask returns the netmask
func (p *IPPool) GetNetmask() string {
	return IPPoolNetmask
}

// BuildKernelIPParam builds the kernel ip= parameter for static IP configuration
// Format: ip=<client-ip>:<server-ip>:<gw-ip>:<netmask>:<hostname>:<device>:<autoconf>
func (p *IPPool) BuildKernelIPParam(ip, hostname string) string {
	// ip=10.0.0.5::10.0.0.1:255.255.255.0:sandbox:eth0:off
	return fmt.Sprintf("ip=%s::%s:%s:%s:eth0:off", ip, IPPoolGateway, IPPoolNetmask, hostname)
}

// Available returns number of available IPs
func (p *IPPool) Available() int {
	p.mu.Lock()
	defer p.mu.Unlock()
	return len(p.available)
}

// removeFromAvailable removes a last octet from the available list
func (p *IPPool) removeFromAvailable(lastOctet int) {
	for i, v := range p.available {
		if v == lastOctet {
			p.available = append(p.available[:i], p.available[i+1:]...)
			return
		}
	}
}
