// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package types

// RunnerInfo contains information about a runner instance
type RunnerInfo struct {
	Id       string            `json:"id"`
	Status   string            `json:"status"`
	Metadata map[string]string `json:"metadata,omitempty"`
} //	@name	RunnerInfo

// AddRunnerResponse contains the response from AddRunners operation
type AddRunnerResponse struct {
	JobID    string   `json:"job_id"`
	PodNames []string `json:"pod_names"`
	Message  string   `json:"message"`
} //	@name	AddRunnerResponse
