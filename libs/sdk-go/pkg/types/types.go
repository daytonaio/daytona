// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package types

import "time"

// CodeLanguage
type CodeLanguage string

const (
	CodeLanguagePython     CodeLanguage = "python"
	CodeLanguageJavaScript CodeLanguage = "javascript"
	CodeLanguageTypeScript CodeLanguage = "typescript"
)

// ExperimentalConfig holds experimental feature flags for the Daytona client.
type ExperimentalConfig struct {
	OtelEnabled bool // Enable OpenTelemetry tracing and metrics
}

// DaytonaConfig represents the configuration for the Daytona client.
// When a field is nil, the client will fall back to environment variables or defaults.
type DaytonaConfig struct {
	APIKey         string
	JWTToken       string
	OrganizationID string
	APIUrl         string
	Target         string
	Experimental   *ExperimentalConfig
}

// Resources represents resource allocation for a sandbox.
type Resources struct {
	CPU    int
	GPU    int
	Memory int
	Disk   int
}

// VolumeMount represents a volume mount configuration
type VolumeMount struct {
	VolumeID  string
	MountPath string
	Subpath   *string // Optional subpath within the volume; nil = mount entire volume
}

// SandboxBaseParams contains common parameters for sandbox creation.
type SandboxBaseParams struct {
	Name                string
	User                string
	Language            CodeLanguage
	EnvVars             map[string]string
	Labels              map[string]string
	Public              bool
	AutoStopInterval    *int // nil = no auto-stop, 0 = immediate stop
	AutoArchiveInterval *int // nil = no auto-archive, 0 = immediate archive
	AutoDeleteInterval  *int // nil = no auto-delete, 0 = immediate delete
	Volumes             []VolumeMount
	NetworkBlockAll     bool
	NetworkAllowList    *string
	Ephemeral           bool
}

// SnapshotParams represents parameters for creating a sandbox from a snapshot
type SnapshotParams struct {
	SandboxBaseParams
	Snapshot string
}

// ImageParams represents parameters for creating a sandbox from an image
type ImageParams struct {
	SandboxBaseParams
	Image     any // string or *Image
	Resources *Resources
}

// CreateSnapshotParams represents parameters for creating a snapshot
type CreateSnapshotParams struct {
	Name           string
	Image          any // string or *Image
	Resources      *Resources
	Entrypoint     []string
	SkipValidation *bool
}

// PaginatedSnapshots represents a paginated list of snapshots
type PaginatedSnapshots struct {
	Items      []*Snapshot
	Total      int
	Page       int
	TotalPages int
}

// Volume represents a Daytona volume
type Volume struct {
	ID             string    `json:"id"`
	Name           string    `json:"name"`
	OrganizationID string    `json:"organizationId"`
	State          string    `json:"state"`
	ErrorReason    *string   `json:"errorReason,omitempty"`
	CreatedAt      time.Time `json:"createdAt"`
	UpdatedAt      time.Time `json:"updatedAt"`
	LastUsedAt     time.Time `json:"lastUsedAt,omitempty"`
}

// Snapshot represents a Daytona snapshot
type Snapshot struct {
	ID             string     `json:"id"`
	OrganizationID string     `json:"organizationId,omitempty"`
	General        bool       `json:"general"`
	Name           string     `json:"name"`
	ImageName      string     `json:"imageName,omitempty"`
	State          string     `json:"state"`
	Size           *float64   `json:"size,omitempty"`
	Entrypoint     []string   `json:"entrypoint,omitempty"`
	CPU            int        `json:"cpu"`
	GPU            int        `json:"gpu"`
	Memory         int        `json:"mem"` // API uses "mem" not "memory"
	Disk           int        `json:"disk"`
	ErrorReason    *string    `json:"errorReason,omitempty"` // nil = success, non-nil = error reason if snapshot failed
	SkipValidation bool       `json:"skipValidation"`
	CreatedAt      time.Time  `json:"createdAt"`
	UpdatedAt      time.Time  `json:"updatedAt"`
	LastUsedAt     *time.Time `json:"lastUsedAt,omitempty"`
}

// FileInfo represents file metadata
type FileInfo struct {
	Name         string
	Size         int64
	Mode         string
	ModifiedTime time.Time
	IsDirectory  bool
}

// FileUpload represents a file to upload
type FileUpload struct {
	Source      any // []byte or string (path)
	Destination string
}

// FileDownloadRequest
type FileDownloadRequest struct {
	Source      string
	Destination *string // nil = download to memory (return []byte), non-nil = save to file path
}

// FileDownloadResponse represents a file download response
type FileDownloadResponse struct {
	Source string
	Result any     // []byte or string (path)
	Error  *string // nil = success, non-nil = error message
}

// GitStatus represents git repository status
type GitStatus struct {
	CurrentBranch   string
	Ahead           int
	Behind          int
	BranchPublished bool
	FileStatus      []FileStatus
}

// FileStatus represents the status of a file in git
type FileStatus struct {
	Path   string
	Status string
}

// GitCommitResponse
type GitCommitResponse struct {
	SHA string
}

// CodeRunParams represents parameters for code execution
type CodeRunParams struct {
	Argv []string
	Env  map[string]string
}

// ExecuteResponse represents a command execution response
type ExecuteResponse struct {
	ExitCode  int
	Result    string
	Artifacts *ExecutionArtifacts // nil when no artifacts available
}

// ExecutionArtifacts represents execution output artifacts
type ExecutionArtifacts struct {
	Stdout string
	Charts []Chart
}

// ExecutionResult represents code interpreter execution result
type ExecutionResult struct {
	Stdout string
	Stderr string
	Charts []Chart         // Optional charts from matplotlib
	Error  *ExecutionError // nil = success, non-nil = execution failed
}

// ExecutionError represents a code execution error
type ExecutionError struct {
	Name      string
	Value     string
	Traceback *string // Optional stack trace; nil when not available
}

// OutputMessage represents an output message
type OutputMessage struct {
	Type      string `json:"type"`
	Text      string `json:"text"`
	Name      string `json:"name"`
	Value     string `json:"value"`
	Traceback string `json:"traceback"`
}

// PtySize represents terminal dimensions
type PtySize struct {
	Rows int
	Cols int
}

// PtyResult represents PTY session exit information
type PtyResult struct {
	ExitCode *int    // nil = process still running, non-nil = exit code
	Error    *string // nil = success, non-nil = error message
}

// PtySessionInfo represents PTY session information
type PtySessionInfo struct {
	ID        string
	Active    bool
	CWD       string // Current working directory; may be empty unavailable
	Cols      int
	Rows      int
	ProcessID *int // Process ID; may be nil if unavailable
	CreatedAt time.Time
}

// ScreenshotRegion represents a screenshot region
type ScreenshotRegion struct {
	X      int
	Y      int
	Width  int
	Height int
}

type ScreenshotOptions struct {
	ShowCursor *bool    // nil = default, true = show, false = hide
	Format     *string  // nil = default format (PNG), or "jpeg", "webp", etc.
	Quality    *int     // nil = default quality, 0-100 for JPEG/WebP
	Scale      *float64 // nil = 1.0, scaling factor for the screenshot
}

type ScreenshotResponse struct {
	Image     string // base64-encoded image data
	Width     int
	Height    int
	SizeBytes *int // Size in bytes
}

type LspLanguageID string

const (
	LspLanguagePython     LspLanguageID = "python"
	LspLanguageJavaScript LspLanguageID = "javascript"
	LspLanguageTypeScript LspLanguageID = "typescript"
)

// Position represents a position in a document
type Position struct {
	Line      int // zero-based
	Character int // zero-based
}

type ChartType string

const (
	ChartTypeLine           ChartType = "line"
	ChartTypeScatter        ChartType = "scatter"
	ChartTypeBar            ChartType = "bar"
	ChartTypePie            ChartType = "pie"
	ChartTypeBoxAndWhisker  ChartType = "box_and_whisker"
	ChartTypeCompositeChart ChartType = "composite_chart"
	ChartTypeUnknown        ChartType = "unknown"
)

// Chart represents a chart
type Chart struct {
	Type     ChartType
	Title    string
	Elements any
	PNG      *string // Optional base64-encoded PNG representation
}
