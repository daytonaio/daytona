// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package docker

import (
	"context"
	"errors"

	"github.com/daytonaio/runner/pkg/api/dto"
	"github.com/daytonaio/runner/pkg/common"
)

func (d *DockerClient) UpdateNetworkSettings(ctx context.Context, containerId string, updateNetworkSettingsDto dto.UpdateNetworkSettingsDTO) error {
	info, err := d.ContainerInspect(ctx, containerId)
	if err != nil {
		return err
	}
	containerShortId := info.ID[:12]

	ipAddress := common.GetContainerIpAddress(ctx, info)

	// Return error if container does not have an IP address
	if ipAddress == "" {
		return errors.New("sandbox does not have an IP address")
	}

	if updateNetworkSettingsDto.NetworkBlockAll != nil && *updateNetworkSettingsDto.NetworkBlockAll {
		err = d.netRulesManager.SetNetworkRules(containerShortId, ipAddress, "")
		if err != nil {
			return err
		}
	} else if updateNetworkSettingsDto.NetworkAllowList != nil {
		err = d.netRulesManager.SetNetworkRules(containerShortId, ipAddress, *updateNetworkSettingsDto.NetworkAllowList)
		if err != nil {
			return err
		}
	}

	if updateNetworkSettingsDto.NetworkLimitEgress != nil && *updateNetworkSettingsDto.NetworkLimitEgress {
		err = d.netRulesManager.SetNetworkLimiter(containerShortId, ipAddress)
		if err != nil {
			return err
		}
	}

	return nil
}
