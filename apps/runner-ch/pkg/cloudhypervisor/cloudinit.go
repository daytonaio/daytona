// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package cloudhypervisor

import (
	"context"
	"fmt"
	"path/filepath"

	log "github.com/sirupsen/logrus"
)

// CloudInitConfig holds the configuration for cloud-init
type CloudInitConfig struct {
	IP        string
	Gateway   string
	Netmask   string
	Hostname  string
	DNS       []string
	SSHPubKey string
}

// createCloudInitISO creates a NoCloud ISO with static IP configuration
// This is much faster than DHCP - cloud-init applies config in ~1-2 seconds
func (c *Client) createCloudInitISO(ctx context.Context, sandboxId string, cfg CloudInitConfig) (string, error) {
	sandboxDir := c.getSandboxDir(sandboxId)
	isoPath := filepath.Join(sandboxDir, "cloud-init.iso")

	// Network config (v2 format for netplan)
	// Use match to handle different interface names (eth0, ens3, enp0s2, etc.)
	networkConfig := fmt.Sprintf(`version: 2
ethernets:
  id0:
    match:
      driver: virtio_net
    addresses:
      - %s/%s
    routes:
      - to: default
        via: %s
    nameservers:
      addresses:
        - 8.8.8.8
        - 8.8.4.4
`, cfg.IP, IPPoolCIDR, cfg.Gateway)

	// Meta-data
	metaData := fmt.Sprintf(`instance-id: %s
local-hostname: %s
`, sandboxId, cfg.Hostname)

	// User-data (minimal)
	userData := `#cloud-config
manage_etc_hosts: true
`

	// Create the ISO using genisoimage/mkisofs on the remote host
	// This creates a NoCloud datasource ISO that cloud-init will read
	script := fmt.Sprintf(`
set -e
TMPDIR=$(mktemp -d)
cd "$TMPDIR"

# Write network config
cat > network-config << 'NETEOF'
%s
NETEOF

# Write meta-data
cat > meta-data << 'METAEOF'
%s
METAEOF

# Write user-data
cat > user-data << 'USEREOF'
%s
USEREOF

# Create ISO (try genisoimage first, then mkisofs)
if command -v genisoimage &>/dev/null; then
    genisoimage -output "%s" -volid cidata -joliet -rock user-data meta-data network-config 2>/dev/null
elif command -v mkisofs &>/dev/null; then
    mkisofs -output "%s" -volid cidata -joliet -rock user-data meta-data network-config 2>/dev/null
else
    # Fallback: create raw cloud-init files if no ISO tool
    mkdir -p "$(dirname %s)"
    cp network-config meta-data user-data "$(dirname %s)/"
    echo "WARNING: No ISO tool available, cloud-init may not work"
fi

rm -rf "$TMPDIR"
`, networkConfig, metaData, userData, isoPath, isoPath, isoPath, isoPath)

	if _, err := c.runCommandOutput(ctx, script); err != nil {
		return "", fmt.Errorf("failed to create cloud-init ISO: %w", err)
	}

	log.Debugf("Created cloud-init ISO at %s with IP %s", isoPath, cfg.IP)
	return isoPath, nil
}

// createCloudInitDisk creates cloud-init files directly (fallback if ISO creation fails)
func (c *Client) createCloudInitDisk(ctx context.Context, sandboxId string, cfg CloudInitConfig) (string, error) {
	sandboxDir := c.getSandboxDir(sandboxId)
	seedDir := filepath.Join(sandboxDir, "seed")

	// Network config (v2 format) - match virtio driver for any interface name
	networkConfig := fmt.Sprintf(`version: 2
ethernets:
  id0:
    match:
      driver: virtio_net
    addresses:
      - %s/%s
    routes:
      - to: default
        via: %s
    nameservers:
      addresses:
        - 8.8.8.8
        - 8.8.4.4
`, cfg.IP, IPPoolCIDR, cfg.Gateway)

	metaData := fmt.Sprintf(`instance-id: %s
local-hostname: %s
`, sandboxId, cfg.Hostname)

	userData := `#cloud-config
manage_etc_hosts: true
`

	// Create seed directory and files
	script := fmt.Sprintf(`
mkdir -p %s
cat > %s/network-config << 'EOF'
%s
EOF
cat > %s/meta-data << 'EOF'
%s
EOF
cat > %s/user-data << 'EOF'
%s
EOF
`, seedDir, seedDir, networkConfig, seedDir, metaData, seedDir, userData)

	if _, err := c.runCommandOutput(ctx, script); err != nil {
		return "", fmt.Errorf("failed to create cloud-init seed: %w", err)
	}

	return seedDir, nil
}
