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
	// Network namespace constants
	NetNSPrefix = "ns-"

	// Fixed internal network (same for all VMs)
	GuestIP      = "192.168.0.2"
	GuestGateway = "192.168.0.1"
	GuestNetmask = "255.255.255.0"
	GuestCIDR    = "24"

	// External network pool: 10.0.{num}.0/24 per sandbox
	// num ranges from 1 to 254
	ExternalNetStart = 1
	ExternalNetEnd   = 254
	ExternalNetBase  = "10.0"
)

// NetNamespace represents a network namespace for a sandbox
type NetNamespace struct {
	SandboxId     string // Sandbox ID
	NamespaceName string // ns-{sandboxId} (truncated to fit)
	VethHost      string // veth-{sandboxId} (host side)
	VethNS        string // veth-{sandboxId}-ns (namespace side)
	TapName       string // tap0 (always same inside namespace)
	GuestIP       string // 192.168.0.2 (always same)
	GatewayIP     string // 192.168.0.1 (always same)
	ExternalNum   int    // Unique number for 10.0.{num}.0/24
	ExternalIP    string // 10.0.{num}.1 (namespace side of veth)
	HostIP        string // 10.0.{num}.254 (host side of veth)
}

// NetNSPool manages network namespaces for sandboxes
type NetNSPool struct {
	mu           sync.Mutex
	client       *Client
	namespaces   map[string]*NetNamespace // sandboxId -> namespace
	availableNum []int                    // available external network numbers
}

// NewNetNSPool creates a new network namespace pool
func NewNetNSPool(client *Client) *NetNSPool {
	pool := &NetNSPool{
		client:       client,
		namespaces:   make(map[string]*NetNamespace),
		availableNum: make([]int, 0, ExternalNetEnd-ExternalNetStart+1),
	}

	// Initialize available pool (1-254)
	for i := ExternalNetStart; i <= ExternalNetEnd; i++ {
		pool.availableNum = append(pool.availableNum, i)
	}

	return pool
}

// Initialize loads existing namespace allocations from sandbox directories
func (p *NetNSPool) Initialize(ctx context.Context) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	// Get list of existing sandboxes
	sandboxes, err := p.client.List(ctx)
	if err != nil {
		return fmt.Errorf("failed to list sandboxes: %w", err)
	}

	log.Infof("NetNS pool: loading allocations for %d existing sandboxes", len(sandboxes))

	for _, sandboxId := range sandboxes {
		// Check if sandbox has stored external network number
		nsFilePath := filepath.Join(p.client.config.SandboxesPath, sandboxId, "netns")
		output, err := p.client.runShellScript(ctx, fmt.Sprintf("cat %s 2>/dev/null", nsFilePath))
		if err != nil {
			continue
		}

		numStr := strings.TrimSpace(output)
		if numStr == "" {
			continue
		}

		num, err := strconv.Atoi(numStr)
		if err != nil {
			continue
		}

		// Restore namespace info
		ns := p.buildNetNamespace(sandboxId, num)
		p.namespaces[sandboxId] = ns
		p.removeFromAvailable(num)
		log.Debugf("NetNS pool: restored allocation %s -> %d", sandboxId, num)
	}

	log.Infof("NetNS pool: %d allocated, %d available", len(p.namespaces), len(p.availableNum))
	return nil
}

// Create creates a new network namespace for a sandbox
func (p *NetNSPool) Create(ctx context.Context, sandboxId string) (*NetNamespace, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	// Check if already exists
	if ns, exists := p.namespaces[sandboxId]; exists {
		return ns, nil
	}

	// Get next available number
	if len(p.availableNum) == 0 {
		return nil, fmt.Errorf("network namespace pool exhausted")
	}

	num := p.availableNum[0]
	p.availableNum = p.availableNum[1:]

	ns := p.buildNetNamespace(sandboxId, num)

	// Create the namespace and network interfaces
	if err := p.createNetworkNamespace(ctx, ns); err != nil {
		// Return number to pool on failure
		p.availableNum = append(p.availableNum, num)
		return nil, fmt.Errorf("failed to create network namespace: %w", err)
	}

	p.namespaces[sandboxId] = ns

	// Store the number for recovery
	nsFilePath := filepath.Join(p.client.config.SandboxesPath, sandboxId, "netns")
	_ = p.client.runCommand(ctx, "sh", "-c", fmt.Sprintf("echo '%d' > %s", num, nsFilePath))

	log.Infof("NetNS pool: created namespace for %s (external: %s.%d.0/24)", sandboxId, ExternalNetBase, num)
	return ns, nil
}

// Delete removes a network namespace for a sandbox
func (p *NetNSPool) Delete(ctx context.Context, sandboxId string) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	ns, exists := p.namespaces[sandboxId]
	if !exists {
		return nil // Nothing to delete
	}

	// Delete the namespace (this also removes veth pairs automatically)
	if err := p.deleteNetworkNamespace(ctx, ns); err != nil {
		log.Warnf("Failed to delete namespace %s: %v", ns.NamespaceName, err)
	}

	// Return number to pool
	p.availableNum = append(p.availableNum, ns.ExternalNum)
	delete(p.namespaces, sandboxId)

	log.Infof("NetNS pool: deleted namespace for %s", sandboxId)
	return nil
}

// Get returns the namespace for a sandbox
func (p *NetNSPool) Get(sandboxId string) *NetNamespace {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.namespaces[sandboxId]
}

// ExecInNamespace executes a command inside a sandbox's network namespace
func (p *NetNSPool) ExecInNamespace(ctx context.Context, sandboxId string, command string) (string, error) {
	p.mu.Lock()
	ns := p.namespaces[sandboxId]
	p.mu.Unlock()

	if ns == nil {
		return "", fmt.Errorf("namespace not found for sandbox %s", sandboxId)
	}

	cmd := fmt.Sprintf("ip netns exec %s %s", ns.NamespaceName, command)
	return p.client.runShellScript(ctx, cmd)
}

// StartProcessInNamespace starts a background process in a sandbox's namespace
func (p *NetNSPool) StartProcessInNamespace(ctx context.Context, sandboxId string, command string) error {
	p.mu.Lock()
	ns := p.namespaces[sandboxId]
	p.mu.Unlock()

	if ns == nil {
		return fmt.Errorf("namespace not found for sandbox %s", sandboxId)
	}

	// Use nohup to keep process running after SSH disconnects
	cmd := fmt.Sprintf("ip netns exec %s nohup %s &", ns.NamespaceName, command)
	_, err := p.client.runShellScript(ctx, cmd)
	return err
}

// Available returns number of available network numbers
func (p *NetNSPool) Available() int {
	p.mu.Lock()
	defer p.mu.Unlock()
	return len(p.availableNum)
}

// buildNetNamespace creates a NetNamespace struct with computed values
func (p *NetNSPool) buildNetNamespace(sandboxId string, num int) *NetNamespace {
	// Truncate sandbox ID for interface names (Linux limit: 15 chars)
	shortId := sandboxId
	if len(shortId) > 8 {
		shortId = shortId[:8]
	}

	return &NetNamespace{
		SandboxId:     sandboxId,
		NamespaceName: fmt.Sprintf("%s%s", NetNSPrefix, shortId),
		VethHost:      fmt.Sprintf("veth-%s", shortId),
		VethNS:        fmt.Sprintf("veth-%s-n", shortId), // -n suffix (namespace side)
		TapName:       "tap0",
		GuestIP:       GuestIP,
		GatewayIP:     GuestGateway,
		ExternalNum:   num,
		ExternalIP:    fmt.Sprintf("%s.%d.1", ExternalNetBase, num),
		HostIP:        fmt.Sprintf("%s.%d.254", ExternalNetBase, num),
	}
}

// createNetworkNamespace creates the actual namespace and network setup
func (p *NetNSPool) createNetworkNamespace(ctx context.Context, ns *NetNamespace) error {
	// Build the complete setup script
	// This creates: namespace, veth pair, TAP inside namespace, NAT/DNAT rules
	//
	// Network topology:
	//   Host (10.0.{num}.254) <--veth--> Namespace (10.0.{num}.1) <--tap--> VM (192.168.0.2)
	//
	// For host to reach VM:
	//   Host connects to 10.0.{num}.1:port -> Namespace DNATs to 192.168.0.2:port
	//
	script := fmt.Sprintf(`
set -e

# 1. Create namespace
ip netns add %s

# 2. Create veth pair (host <-> namespace)
ip link add %s type veth peer name %s
ip link set %s netns %s

# 3. Create TAP inside namespace (namespace <-> VM)
ip netns exec %s ip tuntap add dev %s mode tap
ip netns exec %s ip link set %s up
ip netns exec %s ip addr add %s/%s dev %s

# 4. Configure namespace-side veth
ip netns exec %s ip link set %s up
ip netns exec %s ip addr add %s/24 dev %s
ip netns exec %s ip route add default via %s

# 5. Enable IP forwarding inside namespace
ip netns exec %s sysctl -w net.ipv4.ip_forward=1 > /dev/null

# 6. NAT inside namespace (VM -> external via veth) - for outbound traffic
ip netns exec %s iptables -t nat -A POSTROUTING -o %s -j MASQUERADE

# 7. DNAT inside namespace (external -> VM) - for inbound traffic from host
# Forward traffic to namespace's veth IP to the VM
ip netns exec %s iptables -t nat -A PREROUTING -i %s -j DNAT --to-destination %s

# 8. Host-side veth setup
ip link set %s up
ip addr add %s/24 dev %s

# 9. Host NAT (namespace -> internet) - add rule if not exists
iptables -t nat -C POSTROUTING -s %s.%d.0/24 -o eth0 -j MASQUERADE 2>/dev/null || \
iptables -t nat -A POSTROUTING -s %s.%d.0/24 -o eth0 -j MASQUERADE

# 10. Enable forwarding on host
sysctl -w net.ipv4.ip_forward=1 > /dev/null

echo "OK"
`,
		// 1. Create namespace
		ns.NamespaceName,
		// 2. Create veth pair
		ns.VethHost, ns.VethNS,
		ns.VethNS, ns.NamespaceName,
		// 3. Create TAP inside namespace
		ns.NamespaceName, ns.TapName,
		ns.NamespaceName, ns.TapName,
		ns.NamespaceName, ns.GatewayIP, GuestCIDR, ns.TapName,
		// 4. Configure namespace-side veth
		ns.NamespaceName, ns.VethNS,
		ns.NamespaceName, ns.ExternalIP, ns.VethNS,
		ns.NamespaceName, ns.HostIP,
		// 5. Enable IP forwarding inside namespace
		ns.NamespaceName,
		// 6. NAT inside namespace (outbound)
		ns.NamespaceName, ns.VethNS,
		// 7. DNAT inside namespace (inbound) - forward to VM
		ns.NamespaceName, ns.VethNS, ns.GuestIP,
		// 8. Host-side veth setup
		ns.VethHost,
		ns.HostIP, ns.VethHost,
		// 9. Host NAT
		ExternalNetBase, ns.ExternalNum,
		ExternalNetBase, ns.ExternalNum,
	)

	output, err := p.client.runShellScript(ctx, script)
	if err != nil {
		// Cleanup on failure
		_ = p.deleteNetworkNamespace(ctx, ns)
		return fmt.Errorf("failed to create namespace: %w (output: %s)", err, output)
	}

	log.Debugf("Created network namespace %s with external %s.%d.0/24", ns.NamespaceName, ExternalNetBase, ns.ExternalNum)
	return nil
}

// deleteNetworkNamespace removes the namespace and cleans up
func (p *NetNSPool) deleteNetworkNamespace(ctx context.Context, ns *NetNamespace) error {
	// Deleting the namespace automatically removes:
	// - All interfaces inside it (TAP, veth-ns side)
	// - The host-side veth is also removed when its peer is deleted
	// We just need to clean up iptables rules

	script := fmt.Sprintf(`
# Remove host NAT rule (ignore errors if not exists)
iptables -t nat -D POSTROUTING -s %s.%d.0/24 -o eth0 -j MASQUERADE 2>/dev/null || true

# Remove host-side veth (may already be gone)
ip link del %s 2>/dev/null || true

# Delete namespace (removes everything inside)
ip netns del %s 2>/dev/null || true

echo "OK"
`,
		ExternalNetBase, ns.ExternalNum,
		ns.VethHost,
		ns.NamespaceName,
	)

	_, err := p.client.runShellScript(ctx, script)
	return err
}

// removeFromAvailable removes a number from the available list
func (p *NetNSPool) removeFromAvailable(num int) {
	for i, v := range p.availableNum {
		if v == num {
			p.availableNum = append(p.availableNum[:i], p.availableNum[i+1:]...)
			return
		}
	}
}

// GetExternalVMIP returns the IP address that the host can use to reach the VM
// This is the guest IP (192.168.0.2) but routed through the namespace's external network
func (ns *NetNamespace) GetExternalVMIP() string {
	// The VM is at 192.168.0.2 inside the namespace
	// From the host, we can reach it via the veth at 10.0.{num}.1 which NATs to 192.168.0.2
	// But for direct routing, we need to go through the namespace
	// The simplest approach: host connects to 10.0.{num}.2 and namespace NATs to 192.168.0.2
	// Actually, let's use 192.168.0.2 as the VM IP and route through namespace
	return ns.GuestIP
}

// GetHostRoutableIP returns the IP the host uses to reach the namespace
func (ns *NetNamespace) GetHostRoutableIP() string {
	// Host can reach the namespace via the veth pair
	// Host side: 10.0.{num}.254
	// Namespace side: 10.0.{num}.1
	// Packets to 192.168.0.2 need to go through the namespace
	return ns.ExternalIP // 10.0.{num}.1 is the namespace's external interface
}
