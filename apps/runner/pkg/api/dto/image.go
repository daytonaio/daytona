// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package dto

import specsgen "github.com/daytonaio/runner/pkg/runner/v2/specs/gen"

type PullSnapshotRequestDTO struct {
	*specsgen.PullSnapshotPayload
} //	@name	PullSnapshotRequestDTO

type BuildSnapshotRequestDTO struct {
	*specsgen.BuildSnapshotPayload
} //	@name	BuildSnapshotRequestDTO

type TagImageRequestDTO struct {
	SourceImage string `json:"sourceImage" validate:"required"`
	TargetImage string `json:"targetImage" validate:"required"`
} //	@name	TagImageRequestDTO

type InspectSnapshotInRegistryRequestDTO struct {
	*specsgen.InspectSnapshotInRegistryPayload
} //	@name	InspectSnapshotInRegistryRequest
