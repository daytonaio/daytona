// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package coderun

// CodeRunRequest - request body for code-run endpoint
type CodeRunRequest struct {
	Code     string            `json:"code" validate:"required"`
	Language string            `json:"language" validate:"required"` // python, javascript, typescript
	Argv     []string          `json:"argv,omitempty"`
	Envs     map[string]string `json:"envs,omitempty"`
	Timeout  *uint32           `json:"timeout,omitempty"`
} //	@name	CodeRunRequest

// CodeRunResponse - response from code-run endpoint
type CodeRunResponse struct {
	ExitCode  int               `json:"exitCode"`
	Result    string            `json:"result"`
	Artifacts *CodeRunArtifacts `json:"artifacts,omitempty"`
} //	@name	CodeRunResponse

// CodeRunArtifacts - artifacts extracted from code execution output
type CodeRunArtifacts struct {
	Charts []map[string]interface{} `json:"charts,omitempty"`
} //	@name	CodeRunArtifacts
