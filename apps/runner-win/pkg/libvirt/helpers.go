// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package libvirt

import "context"

// GetDomainIpAddress extracts the IP address from a DomainInfo
// The IP address is populated by ContainerInspect using network DHCP lease information
func GetDomainIpAddress(ctx context.Context, domain DomainInfo) string {
	return domain.IPAddress
}
