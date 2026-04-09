// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package coderun

type CodeRunRequest struct {
	Code     string            `json:"code" validate:"required"`
	Language string            `json:"language" validate:"required"` // python, javascript, typescript
	Argv     []string          `json:"argv,omitempty"`
	Envs     map[string]string `json:"envs,omitempty"`
	Timeout  *uint32           `json:"timeout,omitempty"`
} // @name CodeRunRequest

type CodeRunResponse struct {
	ExitCode  int               `json:"exitCode"`
	Result    string            `json:"result"`
	Artifacts *CodeRunArtifacts `json:"artifacts,omitempty"`
} // @name CodeRunResponse

type CodeRunArtifacts struct {
	Charts []Chart `json:"charts,omitempty"`
} // @name CodeRunArtifacts

type Chart struct {
	Type        string         `json:"type"`
	Title       string         `json:"title,omitempty"`
	Png         string         `json:"png,omitempty"`
	XLabel      string         `json:"x_label,omitempty"`
	YLabel      string         `json:"y_label,omitempty"`
	XTicks      []float64      `json:"x_ticks,omitempty"`
	YTicks      []float64      `json:"y_ticks,omitempty"`
	XTickLabels []string       `json:"x_tick_labels,omitempty"`
	YTickLabels []string       `json:"y_tick_labels,omitempty"`
	XScale      string         `json:"x_scale,omitempty"`
	YScale      string         `json:"y_scale,omitempty"`
	Elements    []ChartElement `json:"elements"`
} // @name Chart

type ChartElement struct {
	Label         string      `json:"label,omitempty"`
	Points        [][]float64 `json:"points,omitempty"`
	Value         *string     `json:"value,omitempty"`
	Group         string      `json:"group,omitempty"`
	Angle         *float64    `json:"angle,omitempty"`
	Radius        *float64    `json:"radius,omitempty"`
	Min           *float64    `json:"min,omitempty"`
	FirstQuartile *float64    `json:"first_quartile,omitempty"`
	Median        *float64    `json:"median,omitempty"`
	ThirdQuartile *float64    `json:"third_quartile,omitempty"`
	Max           *float64    `json:"max,omitempty"`
	Outliers      []float64   `json:"outliers,omitempty"`
	Type          string      `json:"type,omitempty"`
	Title         string      `json:"title,omitempty"`
	Png           string      `json:"png,omitempty"`
	XLabel        string      `json:"x_label,omitempty"`
	YLabel        string      `json:"y_label,omitempty"`
} // @name ChartElement
