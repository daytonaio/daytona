// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package dto

type DiskDTO struct {
	DiskId    string `json:"diskId"`
	MountPath string `json:"mountPath"`
} //	@name	DiskDTO

type DiskInfoDTO struct {
	Name         string `json:"name"`
	SizeGB       int64  `json:"sizeGB"`
	ActualSizeGB int64  `json:"actualSizeGB"`
	Created      string `json:"created"`
	Modified     string `json:"modified"`
	IsMounted    bool   `json:"isMounted"`
	MountPath    string `json:"mountPath"`
	InS3         bool   `json:"inS3"`
	Checksum     string `json:"checksum"`
} //	@name	DiskInfoDTO
