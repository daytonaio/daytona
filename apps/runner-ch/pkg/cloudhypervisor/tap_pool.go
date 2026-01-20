// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package cloudhypervisor

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
)

// TapPoolConfig configures the TAP interface pool
type TapPoolConfig struct {
	Enabled    bool
	Size       int    // Target pool size
	BridgeName string // Network bridge to attach TAPs to
}

// TapPool manages a pool of pre-created TAP interfaces for fast VM creation
type TapPool struct {
	config    TapPoolConfig
	client    *Client
	available []string          // Available TAP names
	inUse     map[string]string // sandboxId -> tapName
	mu        sync.Mutex
	ctx       context.Context
	cancel    context.CancelFunc
}

// NewTapPool creates a new TAP interface pool
func NewTapPool(client *Client, config TapPoolConfig) *TapPool {
	if config.Size <= 0 {
		config.Size = 10 // Default pool size
	}
	if config.BridgeName == "" {
		config.BridgeName = "br0"
	}

	ctx, cancel := context.WithCancel(context.Background())

	return &TapPool{
		config:    config,
		client:    client,
		available: make([]string, 0, config.Size),
		inUse:     make(map[string]string),
		ctx:       ctx,
		cancel:    cancel,
	}
}

// Start initializes the pool and starts the background replenisher
func (p *TapPool) Start(ctx context.Context) error {
	if !p.config.Enabled {
		log.Info("TAP pool disabled")
		return nil
	}

	log.Infof("Starting TAP pool (size=%d, bridge=%s)", p.config.Size, p.config.BridgeName)

	// Initial pool fill
	if err := p.fill(ctx); err != nil {
		log.Warnf("Failed to fill TAP pool: %v", err)
	}

	// Start background replenisher
	go p.replenisher()

	return nil
}

// Stop shuts down the pool and cleans up TAP interfaces
func (p *TapPool) Stop() {
	p.cancel()

	p.mu.Lock()
	defer p.mu.Unlock()

	// Clean up available TAPs
	for _, tap := range p.available {
		if err := p.deleteTap(context.Background(), tap); err != nil {
			log.Warnf("Failed to delete pool TAP %s: %v", tap, err)
		}
	}
	p.available = nil

	log.Info("TAP pool stopped")
}

// Acquire gets a TAP interface from the pool for a sandbox
// Returns the TAP name. If pool is empty, creates one on-demand.
func (p *TapPool) Acquire(ctx context.Context, sandboxId string) (string, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	// Check if sandbox already has a TAP
	if tap, exists := p.inUse[sandboxId]; exists {
		return tap, nil
	}

	var tapName string

	if len(p.available) > 0 {
		// Get from pool
		tapName = p.available[0]
		p.available = p.available[1:]
		log.Debugf("TAP pool: acquired %s for %s (pool size: %d)", tapName, sandboxId, len(p.available))
	} else {
		// Pool empty, create on-demand
		log.Warnf("TAP pool empty, creating on-demand for %s", sandboxId)
		var err error
		tapName, err = p.createTapForSandbox(ctx, sandboxId)
		if err != nil {
			return "", err
		}
	}

	p.inUse[sandboxId] = tapName
	return tapName, nil
}

// Release returns a TAP interface to the pool or deletes it
func (p *TapPool) Release(ctx context.Context, sandboxId string) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	tapName, exists := p.inUse[sandboxId]
	if !exists {
		return nil // Nothing to release
	}

	delete(p.inUse, sandboxId)

	// If pool is full, delete the TAP
	if len(p.available) >= p.config.Size {
		log.Debugf("TAP pool full, deleting %s", tapName)
		return p.deleteTap(ctx, tapName)
	}

	// Return to pool
	p.available = append(p.available, tapName)
	log.Debugf("TAP pool: released %s (pool size: %d)", tapName, len(p.available))
	return nil
}

// GetTapForSandbox returns the TAP name for a sandbox (if using pool)
func (p *TapPool) GetTapForSandbox(sandboxId string) string {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.inUse[sandboxId]
}

// IsEnabled returns true if the pool is enabled
func (p *TapPool) IsEnabled() bool {
	return p.config.Enabled
}

// fill fills the pool to the target size
func (p *TapPool) fill(ctx context.Context) error {
	p.mu.Lock()
	currentSize := len(p.available)
	targetSize := p.config.Size
	p.mu.Unlock()

	needed := targetSize - currentSize
	if needed <= 0 {
		return nil
	}

	log.Debugf("TAP pool: creating %d interfaces", needed)

	// Create TAPs in batch via single SSH command for speed
	created, err := p.createTapsBatch(ctx, needed)
	if err != nil {
		return err
	}

	p.mu.Lock()
	p.available = append(p.available, created...)
	p.mu.Unlock()

	log.Infof("TAP pool filled: %d available", len(p.available))
	return nil
}

// replenisher runs in background to keep pool filled
func (p *TapPool) replenisher() {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-p.ctx.Done():
			return
		case <-ticker.C:
			p.mu.Lock()
			available := len(p.available)
			target := p.config.Size
			p.mu.Unlock()

			// Replenish if below 50% of target
			if available < target/2 {
				if err := p.fill(p.ctx); err != nil {
					log.Warnf("TAP pool replenish failed: %v", err)
				}
			}
		}
	}
}

// createTapsBatch creates multiple TAP interfaces in a single SSH call
func (p *TapPool) createTapsBatch(ctx context.Context, count int) ([]string, error) {
	if count <= 0 {
		return nil, nil
	}

	// Generate unique TAP names
	names := make([]string, count)
	timestamp := time.Now().UnixNano() % 100000
	for i := 0; i < count; i++ {
		// Format: tap-pNNNNN (p for pool, 5 digits) = 11 chars + "tap-" = 15 chars max
		names[i] = fmt.Sprintf("tap-p%05d", (int(timestamp)+i)%100000)
	}

	// Build batch command
	var cmds []string
	for _, name := range names {
		cmds = append(cmds,
			fmt.Sprintf("ip tuntap add %s mode tap", name),
			fmt.Sprintf("ip link set %s master %s", name, p.config.BridgeName),
			fmt.Sprintf("ip link set %s up", name),
		)
	}

	batchCmd := strings.Join(cmds, " && ")
	log.Debugf("Creating %d TAPs in batch", count)

	if err := p.client.runCommand(ctx, "sh", "-c", batchCmd); err != nil {
		// Try to clean up any that were created
		for _, name := range names {
			_ = p.deleteTap(ctx, name)
		}
		return nil, fmt.Errorf("failed to create TAP batch: %w", err)
	}

	return names, nil
}

// createTapForSandbox creates a TAP with sandbox-specific name (fallback)
func (p *TapPool) createTapForSandbox(ctx context.Context, sandboxId string) (string, error) {
	// Use first 11 chars of sandbox ID
	name := sandboxId
	if len(name) > 11 {
		name = name[:11]
	}
	tapName := fmt.Sprintf("tap-%s", name)

	cmd := fmt.Sprintf("ip tuntap add %s mode tap && ip link set %s master %s && ip link set %s up",
		tapName, tapName, p.config.BridgeName, tapName)

	if err := p.client.runCommand(ctx, "sh", "-c", cmd); err != nil {
		return "", fmt.Errorf("failed to create TAP %s: %w", tapName, err)
	}

	return tapName, nil
}

// deleteTap removes a TAP interface
func (p *TapPool) deleteTap(ctx context.Context, tapName string) error {
	return p.client.runCommand(ctx, "ip", "tuntap", "del", tapName, "mode", "tap")
}

// Stats returns pool statistics
func (p *TapPool) Stats() (available, inUse int) {
	p.mu.Lock()
	defer p.mu.Unlock()
	return len(p.available), len(p.inUse)
}
