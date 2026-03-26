// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package port

type PortList struct {
	Ports []uint `json:"ports"`
} // @name PortList

type IsPortInUseResponse struct {
	IsInUse bool `json:"isInUse"`
} // @name IsPortInUseResponse
