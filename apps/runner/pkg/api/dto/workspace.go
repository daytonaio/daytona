// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package dto

type CreateSandboxDTO struct {
	Id           string            `json:"id" validate:"required"`
	FromVolumeId string            `json:"fromVolumeId,omitempty"`
	UserId       string            `json:"userId" validate:"required"`
	Snapshot     string            `json:"snapshot" validate:"required"`
	OsUser       string            `json:"osUser" validate:"required"`
	CpuQuota     int64             `json:"cpuQuota" validate:"min=1"`
	GpuQuota     int64             `json:"gpuQuota" validate:"min=0"`
	MemoryQuota  int64             `json:"memoryQuota" validate:"min=1"`
	StorageQuota int64             `json:"storageQuota" validate:"min=1"`
	Env          map[string]string `json:"env,omitempty"`
	Registry     *RegistryDTO      `json:"registry,omitempty"`
	Entrypoint   []string          `json:"entrypoint,omitempty"`
	Volumes      []VolumeDTO       `json:"volumes,omitempty"`
} //	@name	CreateSandboxDTO

type ResizeSandboxDTO struct {
	Cpu    int64 `json:"cpu" validate:"min=1"`
	Gpu    int64 `json:"gpu" validate:"min=0"`
	Memory int64 `json:"memory" validate:"min=1"`
} //	@name	ResizeSandboxDTO
