// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package docker

import (
	"context"
	"fmt"
	"strings"

	"github.com/containerd/errdefs"
	common_errors "github.com/daytonaio/common-go/pkg/errors"
	"github.com/daytonaio/runner/pkg/api/dto"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	"github.com/vishvananda/netlink"
)

const (
	// linkNetworkPrefix is prepended to the owner sandbox id to form the per-owner
	// link-network name. Presence of a network with this prefix is the single source
	// of truth for "this sandbox owns a link network" during Destroy.
	linkNetworkPrefix = "daytona-link-"

	// sandboxNameLabel carries the human-readable sandbox name on the container.
	// It is used as the DNS alias when attaching to a link network so other linked
	// sandboxes can resolve each other by name.
	sandboxNameLabel = "daytona.sandbox_name"

	// linkOwnerLabel is attached to the link network itself (not the container)
	// and records which sandbox id owns the network.
	linkOwnerLabel = "daytona.link_network_owner"
)

// linkNetworkName returns the per-owner link-network name for the given owner sandbox id.
func linkNetworkName(ownerId string) string {
	return linkNetworkPrefix + ownerId
}

// networkAliasForOwner resolves the alias to use when attaching the owner container
// to its link network. Prefers the daytona.sandbox_name label; falls back to the
// sandbox id for containers created before that label existed.
func networkAliasForOwner(owner *container.InspectResponse) string {
	if owner != nil && owner.Config != nil {
		if name := owner.Config.Labels[sandboxNameLabel]; name != "" {
			return name
		}
	}
	if owner != nil {
		if trimmed := strings.TrimPrefix(owner.Name, "/"); trimmed != "" {
			return trimmed
		}
	}
	return ""
}

// ensureLinkNetwork creates the per-owner link network if it does not yet exist
// and ensures the matching iptables ACCEPT rule is installed so traffic between
// containers on this bridge bypasses the daemon-level ICC block (relevant when
// /etc/docker/daemon.json sets `"icc": false` — the per-network enable_icc flag
// alone does not override the daemon default).
func (d *DockerClient) ensureLinkNetwork(ctx context.Context, ownerId string) error {
	name := linkNetworkName(ownerId)

	existing, err := d.apiClient.NetworkInspect(ctx, name, network.InspectOptions{})
	if err == nil {
		return d.ensureLinkNetworkICCRule(ctx, &existing)
	} else if !errdefs.IsNotFound(err) {
		return fmt.Errorf("failed to inspect link network %s: %w", name, err)
	}

	if _, err := d.apiClient.NetworkCreate(ctx, name, network.CreateOptions{
		Driver: "bridge",
		Options: map[string]string{
			"com.docker.network.bridge.enable_icc": "true",
		},
		Labels: map[string]string{
			linkOwnerLabel: ownerId,
		},
	}); err != nil {
		// If two concurrent followers race to create the network, one will win
		// and the other will see "already exists" — treat that as success.
		if !isAlreadyExistsErr(err) {
			return fmt.Errorf("failed to create link network %s: %w", name, err)
		}
	}

	created, err := d.apiClient.NetworkInspect(ctx, name, network.InspectOptions{})
	if err != nil {
		return fmt.Errorf("failed to inspect link network %s after create: %w", name, err)
	}
	return d.ensureLinkNetworkICCRule(ctx, &created)
}

// ensureLinkNetworkICCRule installs DOCKER-USER ACCEPT rules for the link
// network so containers attached to it can talk to each other even when the
// Docker daemon is configured with `icc: false`. We install two rules — one
// matching by bridge interface (mirroring exactly Docker's own ICC DROP) and
// one matching by bridge + subnet — because they catch slightly different
// traffic shapes and the second rule alone has been observed to be
// insufficient on some Docker versions where the DROP lives in a chain that
// is not bypassed by a subnet-only ACCEPT.
//
// Both rules are anchored to the input interface (`-i <bridge>`). Without
// that anchor, the subnet rule would also match cross-bridge packets with
// spoofed src/dst, letting a privileged container on any other bridge inject
// one-way traffic into the link network and bypass Docker's inter-bridge
// isolation.
//
// No-op when the netrules manager is unset or the network has no usable
// bridge name / IPv4 IPAM config.
func (d *DockerClient) ensureLinkNetworkICCRule(ctx context.Context, net *network.Inspect) error {
	if d.netRulesManager == nil || net == nil {
		return nil
	}

	bridge := linkNetworkBridgeName(net)
	subnet := linkNetworkIPv4Subnet(net)

	if bridge != "" {
		if err := d.netRulesManager.AllowBridgeICC(bridge); err != nil {
			return fmt.Errorf("failed to install ICC bridge rule for link network %s (%s): %w", net.Name, bridge, err)
		}
	}

	if bridge != "" && subnet != "" {
		if err := d.netRulesManager.AllowSubnetICC(bridge, subnet); err != nil {
			return fmt.Errorf("failed to install ICC subnet rule for link network %s (%s/%s): %w", net.Name, bridge, subnet, err)
		}
	}

	d.logger.DebugContext(ctx, "Installed link-network ICC rules",
		"network", net.Name, "bridge", bridge, "subnet", subnet)
	return nil
}

// linkNetworkIPv4Subnet picks the first IPv4 subnet from the network's IPAM
// config. Link networks are created without an explicit subnet, so Docker
// auto-assigns exactly one — but we still loop and skip IPv6 entries to be
// safe against future config changes.
func linkNetworkIPv4Subnet(net *network.Inspect) string {
	if net == nil {
		return ""
	}
	for _, cfg := range net.IPAM.Config {
		if cfg.Subnet == "" {
			continue
		}
		if strings.Contains(cfg.Subnet, ":") {
			continue
		}
		return cfg.Subnet
	}
	return ""
}

// linkNetworkBridgeName returns the host-side Linux bridge interface name for
// a Docker bridge network. If the user explicitly set
// `com.docker.network.bridge.name` we honor it; otherwise we use Docker's
// documented default of `br-<networkID[:12]>` (the 12-char prefix keeps the
// total within Linux's 15-char interface name limit).
func linkNetworkBridgeName(net *network.Inspect) string {
	if net == nil {
		return ""
	}
	if name, ok := net.Options["com.docker.network.bridge.name"]; ok && name != "" {
		return name
	}
	if len(net.ID) < 12 {
		return ""
	}
	return "br-" + net.ID[:12]
}

// clearLinkNetworkIsolation removes the bridge-port `isolated` flag that Docker
// sets on every veth attached to a link network when the daemon is configured
// with `icc: false`. Without this, two sandboxes on the same link bridge cannot
// even ARP for each other (port isolation drops the traffic at L2, before any
// iptables hook fires) — so the per-network enable_icc=true override and the
// DOCKER-USER ACCEPT rules we add are both insufficient on their own.
//
// Best-effort: failures are logged but do not abort the calling Create flow,
// since a failed clear leaves the network in the same broken state we'd be in
// without this code at all and we'd rather surface the original Create error.
func (d *DockerClient) clearLinkNetworkIsolation(ctx context.Context, ownerId string) {
	if ownerId == "" {
		return
	}
	name := linkNetworkName(ownerId)
	net, err := d.apiClient.NetworkInspect(ctx, name, network.InspectOptions{})
	if err != nil {
		if !errdefs.IsNotFound(err) {
			d.logger.WarnContext(ctx, "Failed to inspect link network for isolation clear",
				"network", name, "error", err)
		}
		return
	}
	bridge := linkNetworkBridgeName(&net)
	if bridge == "" {
		return
	}
	if err := d.clearBridgePortIsolation(ctx, bridge); err != nil {
		d.logger.WarnContext(ctx, "Failed to clear bridge port isolation",
			"bridge", bridge, "error", err)
	}
}

// clearBridgePortIsolation walks every link whose master is the named bridge
// and clears `isolated` on it. Idempotent — already-unisolated ports are a
// no-op. Per-port errors are logged and the sweep continues so a single
// transient netlink error doesn't leave half the bridge stuck isolated.
func (d *DockerClient) clearBridgePortIsolation(ctx context.Context, bridgeName string) error {
	bridge, err := netlink.LinkByName(bridgeName)
	if err != nil {
		return fmt.Errorf("look up bridge %s: %w", bridgeName, err)
	}
	links, err := netlink.LinkList()
	if err != nil {
		return fmt.Errorf("list links: %w", err)
	}
	bridgeIdx := bridge.Attrs().Index
	for _, link := range links {
		attrs := link.Attrs()
		if attrs.MasterIndex != bridgeIdx {
			continue
		}
		if err := netlink.LinkSetIsolated(link, false); err != nil {
			d.logger.WarnContext(ctx, "Failed to clear isolation on bridge port",
				"bridge", bridgeName, "port", attrs.Name, "error", err)
			continue
		}
	}
	return nil
}

// ensureOwnerOnLinkNetwork connects the owner container to its link network with an
// alias equal to the sandbox name (falling back to sandbox id). It is a no-op if the
// owner is already attached.
func (d *DockerClient) ensureOwnerOnLinkNetwork(ctx context.Context, owner *container.InspectResponse, ownerId string) error {
	if owner == nil {
		return fmt.Errorf("owner container inspect result is nil")
	}

	name := linkNetworkName(ownerId)
	net, err := d.apiClient.NetworkInspect(ctx, name, network.InspectOptions{})
	if err != nil {
		return fmt.Errorf("failed to inspect link network %s: %w", name, err)
	}

	if _, attached := net.Containers[owner.ID]; attached {
		return nil
	}

	alias := networkAliasForOwner(owner)
	aliases := []string{}
	if alias != "" && alias != ownerId {
		aliases = append(aliases, alias)
	}

	if err := d.apiClient.NetworkConnect(ctx, name, owner.ID, &network.EndpointSettings{
		Aliases: aliases,
	}); err != nil {
		if isAlreadyAttachedErr(err) {
			return nil
		}
		return fmt.Errorf("failed to connect owner %s to link network %s: %w", ownerId, name, err)
	}
	d.clearLinkNetworkIsolation(ctx, ownerId)
	return nil
}

// connectFollowerToLinkNetwork connects the follower container `sandboxId` to the
// owner's link network with an alias of `sandboxName` (if non-empty). Idempotent:
// returns nil if the follower is already attached.
func (d *DockerClient) connectFollowerToLinkNetwork(ctx context.Context, ownerId, sandboxId, sandboxName string) error {
	name := linkNetworkName(ownerId)

	aliases := []string{}
	if sandboxName != "" && sandboxName != sandboxId {
		aliases = append(aliases, sandboxName)
	}

	err := d.apiClient.NetworkConnect(ctx, name, sandboxId, &network.EndpointSettings{
		Aliases: aliases,
	})
	if err != nil && !isAlreadyAttachedErr(err) {
		return fmt.Errorf("failed to connect follower %s to link network %s: %w", sandboxId, name, err)
	}
	d.clearLinkNetworkIsolation(ctx, ownerId)
	return nil
}

// teardownOwnedLinkNetwork removes the link network owned by `sandboxId`, first
// disconnecting any remaining containers still attached to it. No-op when the
// sandbox does not own a link network. Per-operation "not found" errors are
// swallowed so destroy remains idempotent.
func (d *DockerClient) teardownOwnedLinkNetwork(ctx context.Context, sandboxId string) error {
	name := linkNetworkName(sandboxId)
	net, err := d.apiClient.NetworkInspect(ctx, name, network.InspectOptions{})
	if err != nil {
		if errdefs.IsNotFound(err) {
			return nil
		}
		return fmt.Errorf("failed to inspect link network %s: %w", name, err)
	}

	for containerId := range net.Containers {
		if discErr := d.apiClient.NetworkDisconnect(ctx, name, containerId, true); discErr != nil && !errdefs.IsNotFound(discErr) {
			d.logger.WarnContext(ctx, "Failed to disconnect container from link network",
				"network", name, "containerId", containerId, "error", discErr)
		}
	}

	subnet := linkNetworkIPv4Subnet(&net)
	bridge := linkNetworkBridgeName(&net)

	if err := d.apiClient.NetworkRemove(ctx, name); err != nil && !errdefs.IsNotFound(err) {
		return fmt.Errorf("failed to remove link network %s: %w", name, err)
	}

	// Remove the iptables ACCEPT rules installed by ensureLinkNetwork. Done after
	// NetworkRemove so the rules keep allowing traffic for any in-flight container
	// disconnects, and so we don't leave dangling rules if NetworkRemove fails.
	if d.netRulesManager != nil {
		if bridge != "" {
			if err := d.netRulesManager.RemoveBridgeICC(bridge); err != nil {
				d.logger.WarnContext(ctx, "Failed to remove link-network bridge ICC rule",
					"network", name, "bridge", bridge, "error", err)
			}
		}
		if subnet != "" {
			if err := d.netRulesManager.RemoveSubnetICC(bridge, subnet); err != nil {
				d.logger.WarnContext(ctx, "Failed to remove link-network subnet ICC rule",
					"network", name, "bridge", bridge, "subnet", subnet, "error", err)
			}
		}
	}
	return nil
}

// prepareLinkedSandboxNetwork runs the owner-side setup before the follower is created:
// validates the owner exists locally, ensures the per-owner link network exists, and
// ensures the owner is attached to it. It is a no-op (returning "", nil) when the DTO
// does not request linking. Returns the owner sandbox id on success.
func (d *DockerClient) prepareLinkedSandboxNetwork(ctx context.Context, sandboxDto dto.CreateSandboxDTO) (string, error) {
	if sandboxDto.LinkedSandboxId == nil || *sandboxDto.LinkedSandboxId == "" {
		return "", nil
	}
	ownerId := *sandboxDto.LinkedSandboxId

	owner, err := d.ContainerInspect(ctx, ownerId)
	if err != nil {
		if common_errors.IsNotFoundError(err) || errdefs.IsNotFound(err) {
			return "", common_errors.NewBadRequestError(fmt.Errorf("linked sandbox %s is not on this runner", ownerId))
		}
		return "", fmt.Errorf("failed to inspect linked sandbox %s: %w", ownerId, err)
	}

	if err := d.ensureLinkNetwork(ctx, ownerId); err != nil {
		return "", err
	}
	if err := d.ensureOwnerOnLinkNetwork(ctx, owner, ownerId); err != nil {
		return "", err
	}
	return ownerId, nil
}

// reconcileFollowerLinkNetwork is the single entry point Create must go through before
// handing a linked follower off to Start. It ensures the owner's link network exists,
// that the owner is attached to it
//
// Calling this on every Create branch — including the Started/Starting and Stopped/
// Creating retry branches — guarantees that a follower whose first Create attempt
// crashed after ContainerCreate but before NetworkConnect is repaired on retry instead
// of silently starting without ever joining the shared network.
func (d *DockerClient) reconcileFollowerLinkNetwork(ctx context.Context, sandboxDto dto.CreateSandboxDTO) (string, error) {
	ownerId, err := d.prepareLinkedSandboxNetwork(ctx, sandboxDto)
	if err != nil || ownerId == "" {
		return ownerId, err
	}

	if err := d.connectFollowerToLinkNetwork(ctx, ownerId, sandboxDto.Id, sandboxDto.Name); err != nil {
		// Android-device followers pin the link network via HostConfig.NetworkMode at
		// container create time, so an explicit connect here races with docker's own
		// attachment and can surface "container not found" if the container hasn't
		// been created yet, or succeed as a no-op once it has. Non-android followers
		// always need the explicit connect. Swallow not-found for android-device so
		// reconcile is safe to call pre-ContainerCreate; any real failure still bubbles
		// up from the post-create reconcile in the main Create flow.
		if sandboxDto.IsAndroidSandbox() && (errdefs.IsNotFound(err) || common_errors.IsNotFoundError(err)) {
			return ownerId, nil
		}

		return "", err
	}
	return ownerId, nil
}

// isAlreadyExistsErr matches Docker's "network with name X already exists" response
// so concurrent ensureLinkNetwork calls can be treated as success.
func isAlreadyExistsErr(err error) bool {
	if err == nil {
		return false
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "already exists")
}

// isAlreadyAttachedErr matches Docker's "endpoint with name X already exists in network Y"
// so repeated NetworkConnect calls are idempotent.
func isAlreadyAttachedErr(err error) bool {
	if err == nil {
		return false
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "already exists in network") ||
		strings.Contains(msg, "is already attached to network")
}
