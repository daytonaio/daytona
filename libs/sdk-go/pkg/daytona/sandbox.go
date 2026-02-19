// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package daytona

import (
	"context"
	"fmt"
	"time"

	apiclient "github.com/daytonaio/daytona/libs/api-client-go"
	"github.com/daytonaio/daytona/libs/sdk-go/pkg/errors"
	"github.com/daytonaio/daytona/libs/sdk-go/pkg/types"
	"github.com/daytonaio/daytona/libs/toolbox-api-client-go"
)

// Sandbox represents a Daytona sandbox environment.
//
// A Sandbox provides an isolated development environment with file system, git,
// process execution, code interpretation, and desktop automation capabilities.
// Sandboxes can be started, stopped, archived, and deleted.
//
// Access sandbox capabilities through the service fields:
//   - FileSystem: File and directory operations
//   - Git: Git repository operations
//   - Process: Command execution and PTY sessions
//   - CodeInterpreter: Python code execution
//   - ComputerUse: Desktop automation (mouse, keyboard, screenshots)
//
// Example:
//
//	// Create and use a sandbox
//	sandbox, err := client.Create(ctx)
//	if err != nil {
//	    return err
//	}
//	defer sandbox.Delete(ctx)
//
//	// Execute a command
//	result, err := sandbox.Process.ExecuteCommand(ctx, "echo 'Hello'")
//
//	// Work with files
//	err = sandbox.FileSystem.UploadFile(ctx, "local.txt", "/home/user/remote.txt")
type Sandbox struct {
	client        *Client
	otel          *otelState
	ID            string                 // Unique sandbox identifier
	Name          string                 // Human-readable sandbox name
	State         apiclient.SandboxState // Current sandbox state
	Target        string                 // Target region/environment where the sandbox runs
	ToolboxClient *toolbox.APIClient     // Internal API client

	// AutoArchiveInterval is the time in minutes after stopping before auto-archiving.
	// Set to 0 to disable auto-archiving.
	AutoArchiveInterval int

	// AutoDeleteInterval is the time in minutes after stopping before auto-deletion.
	// Set to -1 to disable auto-deletion.
	// Set to 0 to delete immediately upon stopping.
	AutoDeleteInterval int

	// NetworkBlockAll blocks all network access when true.
	NetworkBlockAll bool

	// NetworkAllowList is a comma-separated list of allowed CIDR addresses.
	NetworkAllowList *string

	FileSystem      *FileSystemService      // File system operations
	Git             *GitService             // Git operations
	Process         *ProcessService         // Process and PTY operations
	CodeInterpreter *CodeInterpreterService // Python code execution
	ComputerUse     *ComputerUseService     // Desktop automation
}

// PaginatedSandboxes represents a paginated list of sandboxes.
//
// Deprecated: Use [CursorPaginatedSandboxes] instead.
type PaginatedSandboxes struct {
	Items      []*Sandbox // Sandboxes in this page
	Total      int        // Total number of sandboxes
	Page       int        // Current page number
	TotalPages int        // Total number of pages
}

// ListSandboxesParams contains parameters for listing sandboxes using cursor-based pagination.
type ListSandboxesParams struct {
	Cursor *string                  // Cursor for pagination
	Limit  *int                     // Maximum number of results to return
	States []apiclient.SandboxState // List of states to filter by
}

// CursorPaginatedSandboxes represents a paginated list of sandboxes using cursor-based pagination.
type CursorPaginatedSandboxes struct {
	Items      []*Sandbox // Sandboxes in this page
	NextCursor *string    // Cursor for the next page of results. Nil if there are no more results.
}

// NewSandbox creates a new Sandbox instance.
//
// This is typically called internally by the SDK. Users should create sandboxes
// using [Client.Create] rather than calling this directly.
func NewSandbox(client *Client, toolboxClient *toolbox.APIClient, id string, name string, state apiclient.SandboxState, target string, autoArchiveInterval int, autoDeleteInterval int, networkBlockAll bool, networkAllowList *string) *Sandbox {
	var otelSt *otelState
	if client != nil {
		otelSt = client.Otel
	}
	return &Sandbox{
		client:              client,
		otel:                otelSt,
		ID:                  id,
		Name:                name,
		State:               state,
		Target:              target,
		AutoArchiveInterval: autoArchiveInterval,
		AutoDeleteInterval:  autoDeleteInterval,
		NetworkBlockAll:     networkBlockAll,
		NetworkAllowList:    networkAllowList,
		ToolboxClient:       toolboxClient,
		FileSystem:          NewFileSystemService(toolboxClient, otelSt),
		Git:                 NewGitService(toolboxClient, otelSt),
		Process:             NewProcessService(toolboxClient, otelSt),
		CodeInterpreter:     NewCodeInterpreterService(toolboxClient, otelSt),
		ComputerUse:         NewComputerUseService(toolboxClient, otelSt),
	}
}

// RefreshData refreshes the sandbox data from the API.
//
// This updates the sandbox's State and other properties from the server.
// Useful for checking if the sandbox state has changed.
//
// Example:
//
//	err := sandbox.RefreshData(ctx)
//	if err != nil {
//	    return err
//	}
//	fmt.Printf("Current state: %s\n", sandbox.State)
func (s *Sandbox) RefreshData(ctx context.Context) error {
	return withInstrumentationVoid(ctx, s.otel, "Sandbox", "RefreshData", func(ctx context.Context) error {
		return s.doRefreshData(ctx)
	})
}

func (s *Sandbox) doRefreshData(ctx context.Context) error {
	authCtx := s.client.getAuthContext(ctx)
	sandboxResp, httpResp, err := s.client.apiClient.SandboxAPI.GetSandbox(authCtx, s.ID).Execute()
	if err != nil {
		return errors.ConvertAPIError(err, httpResp)
	}

	// Update sandboxDTO for backward compatibility
	s.Name = sandboxResp.GetName()
	s.State = sandboxResp.GetState()
	s.Target = sandboxResp.GetTarget()

	// Update intervals
	if sandboxResp.AutoArchiveInterval != nil {
		s.AutoArchiveInterval = int(*sandboxResp.AutoArchiveInterval)
	}
	if sandboxResp.AutoDeleteInterval != nil {
		s.AutoDeleteInterval = int(*sandboxResp.AutoDeleteInterval)
	}

	// Update network settings
	s.NetworkBlockAll = sandboxResp.GetNetworkBlockAll()
	s.NetworkAllowList = sandboxResp.NetworkAllowList

	return nil
}

// GetUserHomeDir returns the user's home directory path in the sandbox.
//
// Example:
//
//	homeDir, err := sandbox.GetUserHomeDir(ctx)
//	if err != nil {
//	    return err
//	}
//	fmt.Printf("Home directory: %s\n", homeDir) // e.g., "/home/daytona"
func (s *Sandbox) GetUserHomeDir(ctx context.Context) (string, error) {
	return withInstrumentation(ctx, s.otel, "Sandbox", "GetUserHomeDir", func(ctx context.Context) (string, error) {
		resp, httpResp, err := s.ToolboxClient.InfoAPI.GetUserHomeDir(ctx).Execute()
		if err != nil {
			return "", errors.ConvertToolboxError(err, httpResp)
		}

		return resp.GetDir(), nil
	})
}

// GetWorkingDir returns the current working directory in the sandbox.
//
// Example:
//
//	workDir, err := sandbox.GetWorkingDir(ctx)
//	if err != nil {
//	    return err
//	}
//	fmt.Printf("Working directory: %s\n", workDir)
func (s *Sandbox) GetWorkingDir(ctx context.Context) (string, error) {
	return withInstrumentation(ctx, s.otel, "Sandbox", "GetWorkingDir", func(ctx context.Context) (string, error) {
		resp, httpResp, err := s.ToolboxClient.InfoAPI.GetWorkDir(ctx).Execute()
		if err != nil {
			return "", errors.ConvertToolboxError(err, httpResp)
		}

		return resp.GetDir(), nil
	})
}

// Start starts the sandbox with a default timeout of 60 seconds.
//
// If the sandbox is already running, this is a no-op.
// For custom timeout, use [Sandbox.StartWithTimeout].
//
// Example:
//
//	err := sandbox.Start(ctx)
//	if err != nil {
//	    return err
//	}
//	// Sandbox is now running
func (s *Sandbox) Start(ctx context.Context) error {
	return withInstrumentationVoid(ctx, s.otel, "Sandbox", "Start", func(ctx context.Context) error {
		return s.StartWithTimeout(ctx, 60*time.Second)
	})
}

// StartWithTimeout starts the sandbox with a custom timeout.
//
// The method blocks until the sandbox reaches the "started" state or the timeout
// is exceeded. 0 means no timeout.
//
// Example:
//
//	err := sandbox.StartWithTimeout(ctx, 2*time.Minute)
//	if err != nil {
//	    return err
//	}
func (s *Sandbox) StartWithTimeout(ctx context.Context, timeout time.Duration) error {
	return withInstrumentationVoid(ctx, s.otel, "Sandbox", "StartWithTimeout", func(ctx context.Context) error {
		return s.doStartWithTimeout(ctx, timeout)
	})
}

func (s *Sandbox) doStartWithTimeout(ctx context.Context, timeout time.Duration) error {
	if timeout < 0 {
		return errors.NewDaytonaError("Timeout must be a non-negative number", 0, nil)
	}

	if timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, timeout)
		defer cancel()
	}

	authCtx := s.client.getAuthContext(ctx)
	_, httpResp, err := s.client.apiClient.SandboxAPI.StartSandbox(authCtx, s.ID).Execute()
	if err != nil {
		return errors.ConvertAPIError(err, httpResp)
	}

	return s.WaitForStart(ctx, timeout)
}

// Stop stops the sandbox with a default timeout of 60 seconds.
//
// Stopping a sandbox preserves its state. Use [Sandbox.Start] to resume.
// For custom timeout, use [Sandbox.StopWithTimeout].
//
// Example:
//
//	err := sandbox.Stop(ctx)
func (s *Sandbox) Stop(ctx context.Context) error {
	return withInstrumentationVoid(ctx, s.otel, "Sandbox", "Stop", func(ctx context.Context) error {
		return s.StopWithTimeout(ctx, 60*time.Second)
	})
}

// StopWithTimeout stops the sandbox with a custom timeout.
//
// The method blocks until the sandbox reaches the "stopped" state or the timeout
// is exceeded. 0 means no timeout.
//
// Example:
//
//	err := sandbox.StopWithTimeout(ctx, 2*time.Minute)
func (s *Sandbox) StopWithTimeout(ctx context.Context, timeout time.Duration) error {
	return withInstrumentationVoid(ctx, s.otel, "Sandbox", "StopWithTimeout", func(ctx context.Context) error {
		return s.doStopWithTimeout(ctx, timeout)
	})
}

func (s *Sandbox) doStopWithTimeout(ctx context.Context, timeout time.Duration) error {
	if timeout < 0 {
		return errors.NewDaytonaError("Timeout must be a non-negative number", 0, nil)
	}

	if timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, timeout)
		defer cancel()
	}

	authCtx := s.client.getAuthContext(ctx)
	_, httpResp, err := s.client.apiClient.SandboxAPI.StopSandbox(authCtx, s.ID).Execute()
	if err != nil {
		return errors.ConvertAPIError(err, httpResp)
	}

	return s.WaitForStop(ctx, timeout)
}

// Delete deletes the sandbox with a default timeout of 60 seconds.
//
// This operation is irreversible. All data in the sandbox will be lost.
// For custom timeout, use [Sandbox.DeleteWithTimeout].
//
// Example:
//
//	err := sandbox.Delete(ctx)
func (s *Sandbox) Delete(ctx context.Context) error {
	return withInstrumentationVoid(ctx, s.otel, "Sandbox", "Delete", func(ctx context.Context) error {
		return s.DeleteWithTimeout(ctx, 60*time.Second)
	})
}

// DeleteWithTimeout deletes the sandbox with a custom timeout. 0 means no timeout.
//
// Example:
//
//	err := sandbox.DeleteWithTimeout(ctx, 2*time.Minute)
func (s *Sandbox) DeleteWithTimeout(ctx context.Context, timeout time.Duration) error {
	return withInstrumentationVoid(ctx, s.otel, "Sandbox", "DeleteWithTimeout", func(ctx context.Context) error {
		return s.doDeleteWithTimeout(ctx, timeout)
	})
}

func (s *Sandbox) doDeleteWithTimeout(ctx context.Context, timeout time.Duration) error {
	if timeout < 0 {
		return errors.NewDaytonaError("Timeout must be a non-negative number", 0, nil)
	}

	if timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, timeout)
		defer cancel()
	}

	authCtx := s.client.getAuthContext(ctx)
	_, httpResp, err := s.client.apiClient.SandboxAPI.DeleteSandbox(authCtx, s.ID).Execute()
	if err != nil {
		return errors.ConvertAPIError(err, httpResp)
	}

	return nil
}

// Archive archives the sandbox, preserving its state in cost-effective storage.
//
// When sandboxes are archived, the entire filesystem state is moved to object
// storage, making it possible to keep sandboxes available for extended periods
// at reduced cost. Use [Sandbox.Start] to unarchive and resume.
//
// Example:
//
//	err := sandbox.Archive(ctx)
//	if err != nil {
//	    return err
//	}
//	// Sandbox is now archived and can be restored later
func (s *Sandbox) Archive(ctx context.Context) error {
	return withInstrumentationVoid(ctx, s.otel, "Sandbox", "Archive", func(ctx context.Context) error {
		return s.doArchive(ctx)
	})
}

func (s *Sandbox) doArchive(ctx context.Context) error {
	authCtx := s.client.getAuthContext(ctx)
	_, httpResp, err := s.client.apiClient.SandboxAPI.ArchiveSandbox(authCtx, s.ID).Execute()
	if err != nil {
		return errors.ConvertAPIError(err, httpResp)
	}

	return s.RefreshData(ctx)
}

// WaitForStart waits for the sandbox to reach the "started" state.
//
// This method polls the sandbox state until it's started, encounters an error
// state, or the timeout is exceeded. 0 means no timeout.
//
// Example:
//
//	err := sandbox.WaitForStart(ctx, 2*time.Minute)
//	if err != nil {
//	    return err
//	}
//	// Sandbox is now running
func (s *Sandbox) WaitForStart(ctx context.Context, timeout time.Duration) error {
	return withInstrumentationVoid(ctx, s.otel, "Sandbox", "WaitForStart", func(ctx context.Context) error {
		return s.doWaitForStart(ctx, timeout)
	})
}

func (s *Sandbox) doWaitForStart(ctx context.Context, timeout time.Duration) error {
	if timeout < 0 {
		return errors.NewDaytonaError("Timeout must be a non-negative number", 0, nil)
	}

	if timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, timeout)
		defer cancel()
	}

	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return errors.NewDaytonaTimeoutError(fmt.Sprintf("Sandbox did not start within %s", timeout))
		case <-ticker.C:
			if err := s.RefreshData(ctx); err != nil {
				return err
			}

			if s.State == apiclient.SANDBOXSTATE_STARTED {
				return nil
			}
			if s.State == apiclient.SANDBOXSTATE_ERROR || s.State == apiclient.SANDBOXSTATE_BUILD_FAILED {
				return errors.NewDaytonaError("Sandbox failed to start", 0, nil)
			}
		}
	}
}

// WaitForStop waits for the sandbox to reach the "stopped" state.
//
// This method polls the sandbox state until it's stopped or the timeout is exceeded.
// 0 means no timeout.
//
// Example:
//
//	err := sandbox.WaitForStop(ctx, 2*time.Minute)
func (s *Sandbox) WaitForStop(ctx context.Context, timeout time.Duration) error {
	return withInstrumentationVoid(ctx, s.otel, "Sandbox", "WaitForStop", func(ctx context.Context) error {
		return s.doWaitForStop(ctx, timeout)
	})
}

func (s *Sandbox) doWaitForStop(ctx context.Context, timeout time.Duration) error {
	if timeout < 0 {
		return errors.NewDaytonaError("Timeout must be a non-negative number", 0, nil)
	}

	if timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, timeout)
		defer cancel()
	}

	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return errors.NewDaytonaTimeoutError(fmt.Sprintf("Sandbox did not stop within %s", timeout))
		case <-ticker.C:
			if err := s.RefreshData(ctx); err != nil {
				return err
			}

			if s.State == apiclient.SANDBOXSTATE_STOPPED {
				return nil
			}
		}
	}
}

// SetLabels sets custom labels on the sandbox.
//
// Labels are key-value pairs that can be used for organization and filtering.
// This replaces all existing labels.
//
// Example:
//
//	err := sandbox.SetLabels(ctx, map[string]string{
//	    "environment": "development",
//	    "team": "backend",
//	    "project": "api-server",
//	})
func (s *Sandbox) SetLabels(ctx context.Context, labels map[string]string) error {
	return withInstrumentationVoid(ctx, s.otel, "Sandbox", "SetLabels", func(ctx context.Context) error {
		return s.doSetLabels(ctx, labels)
	})
}

func (s *Sandbox) doSetLabels(ctx context.Context, labels map[string]string) error {
	sandboxLabels := apiclient.SandboxLabels{
		Labels: labels,
	}

	_, httpResp, err := s.client.apiClient.SandboxAPI.ReplaceLabels(
		s.client.getAuthContext(ctx),
		s.ID,
	).SandboxLabels(sandboxLabels).Execute()

	if err != nil {
		return s.client.handleAPIError(err, httpResp)
	}

	return s.RefreshData(ctx)
}

// GetPreviewLink returns a URL for accessing a port on the sandbox.
//
// The preview URL allows external access to services running on the specified
// port within the sandbox.
//
// Example:
//
//	// Start a web server on port 3000 in the sandbox
//	sandbox.Process.ExecuteCommand(ctx, "python -m http.server 3000 &")
//
//	// Get the preview URL
//	url, err := sandbox.GetPreviewLink(ctx, 3000)
//	if err != nil {
//	    return err
//	}
//	fmt.Printf("Access at: %s\n", url)
func (s *Sandbox) GetPreviewLink(ctx context.Context, port int) (string, error) {
	return withInstrumentation(ctx, s.otel, "Sandbox", "GetPreviewLink", func(ctx context.Context) (string, error) {
		result, httpResp, err := s.client.apiClient.SandboxAPI.GetPortPreviewUrl(
			s.client.getAuthContext(ctx),
			s.ID,
			float32(port),
		).Execute()

		if err != nil {
			return "", s.client.handleAPIError(err, httpResp)
		}

		return result.GetUrl(), nil
	})
}

// SetAutoArchiveInterval sets the auto-archive interval in minutes.
//
// The sandbox will be automatically archived after being stopped for this
// many minutes. Set to 0 to disable auto-archiving (sandbox will never
// auto-archive).
//
// Example:
//
//	// Archive after 30 minutes of being stopped
//	interval := 30
//	err := sandbox.SetAutoArchiveInterval(ctx, &interval)
//
//	// Disable auto-archiving
//	interval := 0
//	err := sandbox.SetAutoArchiveInterval(ctx, &interval)
func (s *Sandbox) SetAutoArchiveInterval(ctx context.Context, intervalMinutes *int) error {
	return withInstrumentationVoid(ctx, s.otel, "Sandbox", "SetAutoArchiveInterval", func(ctx context.Context) error {
		return s.doSetAutoArchiveInterval(ctx, intervalMinutes)
	})
}

func (s *Sandbox) doSetAutoArchiveInterval(ctx context.Context, intervalMinutes *int) error {
	if intervalMinutes == nil {
		return errors.NewDaytonaError("intervalMinutes cannot be nil", 0, nil)
	}

	_, httpResp, err := s.client.apiClient.SandboxAPI.SetAutoArchiveInterval(
		s.client.getAuthContext(ctx),
		s.ID,
		float32(*intervalMinutes),
	).Execute()

	if err != nil {
		return errors.ConvertAPIError(err, httpResp)
	}

	s.AutoArchiveInterval = *intervalMinutes
	return nil
}

// SetAutoDeleteInterval sets the auto-delete interval in minutes.
//
// The sandbox will be automatically deleted after being stopped for this
// many minutes.
//
// Special values:
//   - -1: Disable auto-deletion (sandbox will never auto-delete)
//   - 0: Delete immediately upon stopping
//
// Example:
//
//	// Delete after 60 minutes of being stopped
//	interval := 60
//	err := sandbox.SetAutoDeleteInterval(ctx, &interval)
//
//	// Delete immediately when stopped
//	interval := 0
//	err := sandbox.SetAutoDeleteInterval(ctx, &interval)
//
//	// Never auto-delete
//	interval := -1
//	err := sandbox.SetAutoDeleteInterval(ctx, &interval)
func (s *Sandbox) SetAutoDeleteInterval(ctx context.Context, intervalMinutes *int) error {
	return withInstrumentationVoid(ctx, s.otel, "Sandbox", "SetAutoDeleteInterval", func(ctx context.Context) error {
		return s.doSetAutoDeleteInterval(ctx, intervalMinutes)
	})
}

func (s *Sandbox) doSetAutoDeleteInterval(ctx context.Context, intervalMinutes *int) error {
	if intervalMinutes == nil {
		return errors.NewDaytonaError("intervalMinutes cannot be nil", 0, nil)
	}

	_, httpResp, err := s.client.apiClient.SandboxAPI.SetAutoDeleteInterval(
		s.client.getAuthContext(ctx),
		s.ID,
		float32(*intervalMinutes),
	).Execute()

	if err != nil {
		return errors.ConvertAPIError(err, httpResp)
	}

	s.AutoDeleteInterval = *intervalMinutes
	return nil
}

// Resize resizes the sandbox resources with a default timeout of 60 seconds.
//
// Changes the CPU, memory, or disk allocation for the sandbox. Resizing a started
// sandbox allows increasing CPU and memory. To resize disk or decrease resources,
// the sandbox must be stopped first.
//
// Example:
//
//	// Resize a started sandbox (CPU and memory can be increased)
//	err := sandbox.Resize(ctx, &types.Resources{CPU: 4, Memory: 8})
//
//	// Resize a stopped sandbox (CPU, memory, and disk can be changed)
//	sandbox.Stop(ctx)
//	err := sandbox.Resize(ctx, &types.Resources{CPU: 2, Memory: 4, Disk: 30})
func (s *Sandbox) Resize(ctx context.Context, resources *types.Resources) error {
	return withInstrumentationVoid(ctx, s.otel, "Sandbox", "Resize", func(ctx context.Context) error {
		return s.ResizeWithTimeout(ctx, resources, 60*time.Second)
	})
}

// ResizeWithTimeout resizes the sandbox resources with a custom timeout.
//
// Changes the CPU, memory, or disk allocation for the sandbox. Resizing a started
// sandbox allows increasing CPU and memory. To resize disk or decrease resources,
// the sandbox must be stopped first. 0 means no timeout.
//
// Example:
//
//	err := sandbox.ResizeWithTimeout(ctx, &types.Resources{CPU: 4, Memory: 8}, 2*time.Minute)
func (s *Sandbox) ResizeWithTimeout(ctx context.Context, resources *types.Resources, timeout time.Duration) error {
	return withInstrumentationVoid(ctx, s.otel, "Sandbox", "ResizeWithTimeout", func(ctx context.Context) error {
		return s.doResizeWithTimeout(ctx, resources, timeout)
	})
}

func (s *Sandbox) doResizeWithTimeout(ctx context.Context, resources *types.Resources, timeout time.Duration) error {
	if resources == nil {
		return errors.NewDaytonaError("Resources must not be nil", 0, nil)
	}

	if timeout < 0 {
		return errors.NewDaytonaError("Timeout must be a non-negative number", 0, nil)
	}

	if timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, timeout)
		defer cancel()
	}

	startTime := time.Now()

	resizeRequest := apiclient.NewResizeSandbox()
	if resources.CPU > 0 {
		resizeRequest.SetCpu(int32(resources.CPU))
	}
	if resources.Memory > 0 {
		resizeRequest.SetMemory(int32(resources.Memory))
	}
	if resources.Disk > 0 {
		resizeRequest.SetDisk(int32(resources.Disk))
	}

	authCtx := s.client.getAuthContext(ctx)
	sandboxResp, httpResp, err := s.client.apiClient.SandboxAPI.ResizeSandbox(authCtx, s.ID).ResizeSandbox(*resizeRequest).Execute()
	if err != nil {
		return errors.ConvertAPIError(err, httpResp)
	}

	// Update sandbox data from response
	s.Name = sandboxResp.GetName()
	s.State = sandboxResp.GetState()
	s.Target = sandboxResp.GetTarget()

	var remainingTimeout time.Duration
	if timeout == 0 {
		remainingTimeout = 0
	} else {
		timeElapsed := time.Since(startTime)
		remainingTimeout = timeout - timeElapsed
		if remainingTimeout <= 0 {
			remainingTimeout = time.Millisecond
		}
	}

	return s.WaitForResize(ctx, remainingTimeout)
}

// WaitForResize waits for the sandbox resize operation to complete.
//
// This method polls the sandbox state until it's no longer resizing, encounters an
// error state, or the timeout is exceeded. 0 means no timeout.
//
// Example:
//
//	err := sandbox.WaitForResize(ctx, 2*time.Minute)
func (s *Sandbox) WaitForResize(ctx context.Context, timeout time.Duration) error {
	return withInstrumentationVoid(ctx, s.otel, "Sandbox", "WaitForResize", func(ctx context.Context) error {
		if timeout < 0 {
			return errors.NewDaytonaError("Timeout must be a non-negative number", 0, nil)
		}

		if timeout > 0 {
			var cancel context.CancelFunc
			ctx, cancel = context.WithTimeout(ctx, timeout)
			defer cancel()
		}

		ticker := time.NewTicker(1 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return errors.NewDaytonaTimeoutError(fmt.Sprintf("Sandbox resize did not complete within %s", timeout))
			case <-ticker.C:
				if err := s.RefreshData(ctx); err != nil {
					return err
				}

				if s.State == apiclient.SANDBOXSTATE_ERROR || s.State == apiclient.SANDBOXSTATE_BUILD_FAILED {
					return errors.NewDaytonaError("Sandbox resize failed", 0, nil)
				}
				if s.State != apiclient.SANDBOXSTATE_RESIZING {
					return nil
				}
			}
		}
	})
}
