// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package dto

type CreateSandboxDTO struct {
	Id               string            `json:"id" validate:"required"`
	FromVolumeId     string            `json:"fromVolumeId,omitempty"`
	UserId           string            `json:"userId" validate:"required"`
	Snapshot         string            `json:"snapshot" validate:"required"`
	OsUser           string            `json:"osUser" validate:"required"`
	CpuQuota         int64             `json:"cpuQuota" validate:"min=1"`
	GpuQuota         int64             `json:"gpuQuota" validate:"min=0"`
	MemoryQuota      int64             `json:"memoryQuota" validate:"min=1"`
	StorageQuota     int64             `json:"storageQuota" validate:"min=1"`
	Env              map[string]string `json:"env,omitempty"`
	Registry         *RegistryDTO      `json:"registry,omitempty"`
	Entrypoint       []string          `json:"entrypoint,omitempty"`
	Volumes          []VolumeDTO       `json:"volumes,omitempty"`
	NetworkBlockAll  *bool             `json:"networkBlockAll,omitempty"`
	NetworkAllowList *string           `json:"networkAllowList,omitempty"`
	Metadata         map[string]string `json:"metadata,omitempty"`
	AuthToken        *string           `json:"authToken,omitempty"`
	OtelEndpoint     *string           `json:"otelEndpoint,omitempty"`
	SkipStart        *bool             `json:"skipStart,omitempty"`
} //	@name	CreateSandboxDTO

type ResizeSandboxDTO struct {
	Cpu    int64 `json:"cpu,omitempty" validate:"omitempty,min=1"`
	Gpu    int64 `json:"gpu,omitempty" validate:"omitempty,min=0"`
	Memory int64 `json:"memory,omitempty" validate:"omitempty,min=1"`
	Disk   int64 `json:"disk,omitempty" validate:"omitempty,min=1"`
} //	@name	ResizeSandboxDTO

type UpdateNetworkSettingsDTO struct {
	NetworkBlockAll    *bool   `json:"networkBlockAll,omitempty"`
	NetworkAllowList   *string `json:"networkAllowList,omitempty"`
	NetworkLimitEgress *bool   `json:"networkLimitEgress,omitempty"`
} //	@name	UpdateNetworkSettingsDTO

type RecoverSandboxDTO struct {
	FromVolumeId      string            `json:"fromVolumeId,omitempty"`
	UserId            string            `json:"userId" validate:"required"`
	Snapshot          *string           `json:"snapshot,omitempty"`
	OsUser            string            `json:"osUser" validate:"required"`
	CpuQuota          int64             `json:"cpuQuota" validate:"min=1"`
	GpuQuota          int64             `json:"gpuQuota" validate:"min=0"`
	MemoryQuota       int64             `json:"memoryQuota" validate:"min=1"`
	StorageQuota      int64             `json:"storageQuota" validate:"min=1"`
	Env               map[string]string `json:"env,omitempty"`
	Volumes           []VolumeDTO       `json:"volumes,omitempty"`
	NetworkBlockAll   *bool             `json:"networkBlockAll,omitempty"`
	NetworkAllowList  *string           `json:"networkAllowList,omitempty"`
	ErrorReason       string            `json:"errorReason" validate:"required"`
	BackupErrorReason string            `json:"backupErrorReason,omitempty"`
} //	@name	RecoverSandboxDTO

type IsRecoverableDTO struct {
	ErrorReason string `json:"errorReason" validate:"required"`
} //	@name	IsRecoverableDTO

type IsRecoverableResponse struct {
	Recoverable bool `json:"recoverable"`
} //	@name	IsRecoverableResponse
type StartSandboxResponse struct {
	DaemonVersion string `json:"daemonVersion"`
} //	@name	StartSandboxResponse
