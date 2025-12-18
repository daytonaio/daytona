// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package dto

import "strings"

type SnapshotInfoResponse struct {
	Name       string   `json:"name" example:"nginx:latest"`
	SizeGB     float64  `json:"sizeGB" example:"0.13"`
	Entrypoint []string `json:"entrypoint,omitempty" example:"[\"nginx\",\"-g\",\"daemon off;\"]"`
	Cmd        []string `json:"cmd,omitempty" example:"[\"nginx\",\"-g\",\"daemon off;\"]"`
	Hash       string   `json:"hash,omitempty" example:"a7be6198544f09a75b26e6376459b47c5b9972e7351d440e092c4faa9ea064ff"`
} //	@name	SnapshotInfoResponse

func HashWithoutPrefix(hash string) string {
	return strings.TrimPrefix(hash, "sha256:")
}
