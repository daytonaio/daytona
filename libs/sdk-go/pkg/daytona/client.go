// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

// Package daytona provides a Go SDK for interacting with the Daytona platform.
//
// The Daytona SDK enables developers to programmatically create, manage, and interact
// with sandboxes - isolated development environments that can run code, execute commands,
// and manage files.
//
// # Getting Started
//
// Create a client using your API key or JWT token:
//
//	client, err := daytona.NewClient()
//	if err != nil {
//	    log.Fatal(err)
//	}
//
// The client reads configuration from environment variables:
//   - DAYTONA_API_KEY: API key for authentication
//   - DAYTONA_JWT_TOKEN: JWT token for authentication (alternative to API key)
//   - DAYTONA_ORGANIZATION_ID: Organization ID (required when using JWT token)
//   - DAYTONA_API_URL: API URL (defaults to https://app.daytona.io/api)
//   - DAYTONA_TARGET: Target environment
//
// Or provide configuration explicitly:
//
//	client, err := daytona.NewClientWithConfig(&types.DaytonaConfig{
//	    APIKey: "your-api-key",
//	    APIUrl: "https://your-instance.daytona.io/api",
//	})
//
// # Creating Sandboxes
//
// Create a sandbox from a snapshot:
//
//	sandbox, err := client.Create(ctx, types.SnapshotParams{
//	    Snapshot: "my-snapshot",
//	})
//
// Create a sandbox from a Docker image:
//
//	sandbox, err := client.Create(ctx, types.ImageParams{
//	    Image: "python:3.11",
//	})
//
// # Working with Sandboxes
//
// Execute code in a sandbox:
//
//	result, err := sandbox.Process.CodeRun(ctx, "print('Hello, World!')")
//
// Run shell commands:
//
//	result, err := sandbox.Process.ExecuteCommand(ctx, "ls -la")
package daytona

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	apiclient "github.com/daytonaio/daytona/libs/api-client-go"
	"github.com/daytonaio/daytona/libs/sdk-go/pkg/common"
	"github.com/daytonaio/daytona/libs/sdk-go/pkg/errors"
	"github.com/daytonaio/daytona/libs/sdk-go/pkg/options"
	"github.com/daytonaio/daytona/libs/sdk-go/pkg/types"
	toolbox "github.com/daytonaio/daytona/libs/toolbox-api-client-go"
	cmap "github.com/orcaman/concurrent-map/v2"
)

const (
	// defaultAPIURL is the default Daytona API endpoint used when no custom URL is configured.
	defaultAPIURL = "https://app.daytona.io/api"

	// sdkSource identifies requests as originating from the Go SDK for telemetry purposes.
	sdkSource = "go-sdk"

	// defaultTimeout is the default timeout duration for API requests and sandbox operations.
	defaultTimeout = 60 * time.Second
)

// Client is the main entry point for interacting with the Daytona platform.
//
// Client provides methods to create, retrieve, list, and manage sandboxes. It handles
// authentication, API communication, and provides access to services like Volume and Snapshot
// management.
//
// Create a Client using [NewClient] or [NewClientWithConfig]:
//
//	client, err := daytona.NewClient()
//	if err != nil {
//	    log.Fatal(err)
//	}
//
// The Client is safe for concurrent use by multiple goroutines.
type Client struct {
	apiKey          string             // API key for authentication
	jwtToken        string             // JWT token for authentication (alternative to apiKey)
	organizationID  string             // Organization ID (required when using JWT)
	apiURL          string             // Base URL for the Daytona API
	region          string             // Region for sandbox creation
	httpClient      *http.Client       // HTTP client for API requests
	defaultLanguage types.CodeLanguage // Default programming language for code execution

	// Otel holds OpenTelemetry state; nil when OTel is disabled.
	Otel *otelState

	// toolboxProxyCache caches toolbox proxy URLs per region.
	// Key: region string, Value: proxy URL string
	toolboxProxyCache cmap.ConcurrentMap[string, string]

	// apiClient is the underlying OpenAPI-generated client
	apiClient *apiclient.APIClient

	// Volume provides methods for managing persistent volumes.
	Volume *VolumeService

	// Snapshot provides methods for managing sandbox snapshots.
	Snapshot *SnapshotService
}

// NewClient creates a new Daytona client with default configuration.
//
// NewClient reads configuration from environment variables:
//   - DAYTONA_API_KEY or DAYTONA_JWT_TOKEN for authentication (one is required)
//   - DAYTONA_ORGANIZATION_ID (required when using JWT token)
//   - DAYTONA_API_URL for custom API endpoint
//   - DAYTONA_TARGET for target environment
//
// For explicit configuration, use [NewClientWithConfig] instead.
func NewClient() (*Client, error) {
	return NewClientWithConfig(nil)
}

// NewClientWithConfig creates a new Daytona client with a custom configuration.
//
// Configuration values provided in config take precedence over environment variables.
// Any configuration field left empty will fall back to the corresponding environment
// variable (see [NewClient] for the list of supported variables).
//
// Example:
//
//	client, err := daytona.NewClientWithConfig(&types.DaytonaConfig{
//	    APIKey:         "your-api-key",
//	    APIUrl:         "https://custom.daytona.io/api",
//	    OrganizationID: "org-123",
//	})
//	if err != nil {
//	    log.Fatal(err)
//	}
//
// Returns an error if neither API key nor JWT token is provided, or if JWT token
// is provided without an organization ID.
func NewClientWithConfig(config *types.DaytonaConfig) (*Client, error) {
	client := &Client{
		defaultLanguage: types.CodeLanguagePython,
		httpClient: &http.Client{
			Timeout: defaultTimeout,
		},
	}

	// Set configuration from config or environment variables
	if config != nil {
		if config.APIKey != "" {
			client.apiKey = config.APIKey
		}
		if config.JWTToken != "" {
			client.jwtToken = config.JWTToken
		}
		if config.OrganizationID != "" {
			client.organizationID = config.OrganizationID
		}
		if config.APIUrl != "" {
			client.apiURL = config.APIUrl
		}
		if config.Target != "" {
			client.region = config.Target
		}
	}

	// Load from environment variables if not set
	if client.apiKey == "" {
		client.apiKey = os.Getenv("DAYTONA_API_KEY")
	}

	if client.jwtToken == "" {
		client.jwtToken = os.Getenv("DAYTONA_JWT_TOKEN")
	}

	// Validate authentication
	if client.apiKey == "" && client.jwtToken == "" {
		return nil, errors.NewDaytonaError("API key or JWT token is required", 0, nil)
	}

	if client.organizationID == "" {
		client.organizationID = os.Getenv("DAYTONA_ORGANIZATION_ID")
	}

	if client.jwtToken != "" && client.organizationID == "" {
		return nil, errors.NewDaytonaError("Organization ID is required when using JWT token", 0, nil)
	}

	if client.apiURL == "" {
		client.apiURL = defaultAPIURL
		if apiURL := os.Getenv("DAYTONA_API_URL"); apiURL != "" {
			client.apiURL = apiURL
		} else if serverURL := os.Getenv("DAYTONA_SERVER_URL"); serverURL != "" {
			client.apiURL = serverURL
		}
	}

	if client.region == "" {
		client.region = os.Getenv("DAYTONA_TARGET")
	}

	// Initialize api-client-go
	apiCfg := apiclient.NewConfiguration()
	apiCfg.Host = common.ExtractHost(client.apiURL)
	apiCfg.Scheme = common.ExtractScheme(client.apiURL)
	apiCfg.HTTPClient = client.httpClient
	apiCfg.AddDefaultHeader("X-Daytona-Source", sdkSource)
	apiCfg.AddDefaultHeader("X-Daytona-SDK-Version", Version)

	// Set server URL with base path
	basePath := common.ExtractPath(client.apiURL)
	apiCfg.Servers = apiclient.ServerConfigurations{
		{URL: fmt.Sprintf("%s://%s%s", apiCfg.Scheme, apiCfg.Host, basePath)},
	}

	// Add organization header if using JWT
	if client.jwtToken != "" {
		apiCfg.AddDefaultHeader("X-Daytona-Organization-ID", client.organizationID)
	}

	client.apiClient = apiclient.NewAPIClient(apiCfg)

	// Initialize OpenTelemetry if enabled
	otelEnabled := (config != nil && config.Experimental != nil && config.Experimental.OtelEnabled) || os.Getenv("DAYTONA_EXPERIMENTAL_OTEL_ENABLED") == "true"
	if otelEnabled {
		otelState, err := initOtel(context.Background())
		if err != nil {
			return nil, errors.NewDaytonaError(fmt.Sprintf("failed to initialize OpenTelemetry: %v", err), 0, nil)
		}
		client.Otel = otelState

		// Wrap HTTP transport with trace context propagation
		base := client.httpClient.Transport
		if base == nil {
			base = http.DefaultTransport
		}
		client.httpClient.Transport = &otelTransport{base: base}
	}

	// Initialize services
	client.Volume = NewVolumeService(client)
	client.Snapshot = NewSnapshotService(client)
	client.toolboxProxyCache = cmap.New[string]()

	return client, nil
}

// Close shuts down the client and releases resources.
// When OpenTelemetry is enabled, Close flushes and shuts down the OTel providers.
// It is safe to call Close even when OTel is not enabled.
func (c *Client) Close(ctx context.Context) error {
	return shutdownOtel(ctx, c.Otel)
}

// getAuthContext returns a context with authentication for api-client-go
func (c *Client) getAuthContext(ctx context.Context) context.Context {
	token := c.apiKey
	if token == "" {
		token = c.jwtToken
	}
	return context.WithValue(ctx, apiclient.ContextAccessToken, token)
}

// handleAPIError converts API client errors to Daytona error types
func (c *Client) handleAPIError(err error, httpResp *http.Response) error {
	if httpResp == nil {
		return errors.NewDaytonaError(err.Error(), 0, nil)
	}

	// Extract error message
	message := err.Error()
	if apiErr, ok := err.(apiclient.GenericOpenAPIError); ok {
		if body := apiErr.Body(); len(body) > 0 {
			message = string(body)
		}
	}

	// Map to specific error types based on status code
	switch httpResp.StatusCode {
	case http.StatusNotFound:
		return errors.NewDaytonaNotFoundError(message, httpResp.Header)
	case http.StatusTooManyRequests:
		return errors.NewDaytonaRateLimitError(message, httpResp.Header)
	default:
		return errors.NewDaytonaError(message, httpResp.StatusCode, httpResp.Header)
	}
}

// createToolboxClient creates a configured toolbox client for a specific sandbox.
// The region parameter is used as the key for caching the toolbox proxy URL.
func (c *Client) createToolboxClient(ctx context.Context, sandboxID string, region string) (*toolbox.APIClient, error) {
	proxyURL, err := c.getProxyToolboxURL(ctx, sandboxID, region)
	if err != nil {
		return nil, err
	}

	// Construct full toolbox URL for this sandbox
	toolboxURL := fmt.Sprintf("%s/%s", proxyURL, sandboxID)

	cfg := toolbox.NewConfiguration()
	cfg.Host = common.ExtractHost(toolboxURL)
	cfg.Scheme = common.ExtractScheme(toolboxURL)
	cfg.HTTPClient = c.httpClient

	// Set base path
	basePath := common.ExtractPath(toolboxURL)
	cfg.Servers = toolbox.ServerConfigurations{
		{URL: fmt.Sprintf("%s://%s%s", cfg.Scheme, cfg.Host, basePath)},
	}

	// Add auth headers
	token := c.apiKey
	if token == "" {
		token = c.jwtToken
	}
	cfg.AddDefaultHeader("Authorization", "Bearer "+token)
	cfg.AddDefaultHeader("X-Daytona-Source", sdkSource)
	cfg.AddDefaultHeader("X-Daytona-SDK-Version", Version)

	if c.jwtToken != "" {
		cfg.AddDefaultHeader("X-Daytona-Organization-ID", c.organizationID)
	}

	return toolbox.NewAPIClient(cfg), nil
}

// Create creates a new sandbox with the specified parameters.
//
// The params argument accepts either [types.SnapshotParams] to create from a snapshot,
// or [types.ImageParams] to create from a Docker image:
//
//	// Create from a snapshot
//	sandbox, err := client.Create(ctx, types.SnapshotParams{
//	    Snapshot: "my-snapshot",
//	    SandboxBaseParams: types.SandboxBaseParams{
//	        Name: "my-sandbox",
//	    },
//	})
//
//	// Create from a Docker image
//	sandbox, err := client.Create(ctx, types.ImageParams{
//	    Image: "python:3.11",
//	    Resources: &types.Resources{
//	        CPU:    2,
//	        Memory: 4096,
//	    },
//	})
//
//	// Create with custom base parameters only
//	sandbox, err := client.Create(ctx, types.SandboxBaseParams{
//	    Name:   "my-sandbox",
//	    Labels: map[string]string{"team": "infra"},
//	})
//
// By default, Create waits for the sandbox to reach the started state before returning.
// Use [options.WithWaitForStart](false) to return immediately after creation.
//
// Optional parameters can be configured using functional options:
//   - [options.WithTimeout]: Set maximum wait time for creation
//   - [options.WithWaitForStart]: Control whether to wait for started state
//   - [options.WithLogChannel]: Receive build logs during image builds
//
// Returns the created [Sandbox] or an error if creation fails.
func (c *Client) Create(ctx context.Context, params any, opts ...func(*options.CreateSandbox)) (*Sandbox, error) {
	return withInstrumentation(ctx, c.Otel, "Client", "Create", func(ctx context.Context) (*Sandbox, error) {
		return c.doCreate(ctx, params, opts...)
	})
}

func (c *Client) doCreate(ctx context.Context, params any, opts ...func(*options.CreateSandbox)) (*Sandbox, error) {
	// Apply options with defaults
	createOpts := &options.CreateSandbox{
		WaitForStart: true, // default to true
	}
	for _, opt := range opts {
		opt(createOpts)
	}

	// Extract values from options
	timeoutVal := defaultTimeout
	if createOpts.Timeout != nil {
		timeoutVal = *createOpts.Timeout
	}

	var cancel context.CancelFunc
	ctx, cancel = context.WithTimeout(ctx, timeoutVal)
	defer cancel()

	// Determine params type and set defaults
	var baseParams types.SandboxBaseParams
	var snapshot string
	var image any
	var resources *types.Resources

	switch p := params.(type) {
	case types.SnapshotParams:
		baseParams = p.SandboxBaseParams
		snapshot = p.Snapshot
	case *types.SnapshotParams:
		if p != nil {
			baseParams = p.SandboxBaseParams
			snapshot = p.Snapshot
		}
	case types.ImageParams:
		baseParams = p.SandboxBaseParams
		image = p.Image
		resources = p.Resources
	case *types.ImageParams:
		if p != nil {
			baseParams = p.SandboxBaseParams
			image = p.Image
			resources = p.Resources
		}
	case types.SandboxBaseParams:
		baseParams = p
	case *types.SandboxBaseParams:
		if p != nil {
			baseParams = *p
		}
	default:
		// Default params
		baseParams = types.SandboxBaseParams{
			Language: c.defaultLanguage,
		}
	}

	// Set default language if not provided
	if baseParams.Language == "" {
		baseParams.Language = types.CodeLanguagePython
	}

	// Validate intervals
	if baseParams.AutoStopInterval != nil && *baseParams.AutoStopInterval < 0 {
		return nil, errors.NewDaytonaError("autoStopInterval must be a non-negative integer", 0, nil)
	}
	if baseParams.AutoArchiveInterval != nil && *baseParams.AutoArchiveInterval < 0 {
		return nil, errors.NewDaytonaError("autoArchiveInterval must be a non-negative integer", 0, nil)
	}

	// Handle ephemeral sandboxes
	if baseParams.Ephemeral {
		zero := 0
		baseParams.AutoDeleteInterval = &zero
	}

	// Build CreateSandbox request using api-client-go
	createReq := apiclient.NewCreateSandbox()

	// Set base parameters
	if baseParams.Name != "" {
		createReq.SetName(baseParams.Name)
	}
	if baseParams.User != "" {
		createReq.SetUser(baseParams.User)
	}
	createReq.SetPublic(baseParams.Public)
	createReq.SetNetworkBlockAll(baseParams.NetworkBlockAll)

	if baseParams.EnvVars != nil {
		createReq.SetEnv(baseParams.EnvVars)
	}
	if baseParams.Labels != nil {
		createReq.SetLabels(baseParams.Labels)
	}
	if c.region != "" {
		createReq.SetTarget(c.region)
	}
	if baseParams.AutoStopInterval != nil {
		createReq.SetAutoStopInterval(int32(*baseParams.AutoStopInterval))
	}
	if baseParams.AutoArchiveInterval != nil {
		createReq.SetAutoArchiveInterval(int32(*baseParams.AutoArchiveInterval))
	}
	if baseParams.AutoDeleteInterval != nil {
		createReq.SetAutoDeleteInterval(int32(*baseParams.AutoDeleteInterval))
	}
	// Convert SDK VolumeMount to API SandboxVolume
	if len(baseParams.Volumes) > 0 {
		apiVolumes := make([]apiclient.SandboxVolume, len(baseParams.Volumes))
		for i, vol := range baseParams.Volumes {
			apiVolumes[i] = *apiclient.NewSandboxVolume(vol.VolumeID, vol.MountPath)
			if vol.Subpath != nil {
				apiVolumes[i].SetSubpath(*vol.Subpath)
			}
		}
		createReq.SetVolumes(apiVolumes)
	}
	if baseParams.NetworkAllowList != nil {
		createReq.SetNetworkAllowList(*baseParams.NetworkAllowList)
	}

	// Handle snapshot
	if snapshot != "" {
		createReq.SetSnapshot(snapshot)
	}

	// Handle image and resources
	if image != nil {
		if imageStr, ok := image.(string); ok {
			buildInfo := apiclient.CreateBuildInfo{
				DockerfileContent: fmt.Sprintf("FROM %s", imageStr),
			}
			createReq.SetBuildInfo(buildInfo)
		} else if img, ok := image.(*DockerImage); ok {
			// Process Image builder
			contextHashes, err := c.processImageContext(ctx, img)
			if err != nil {
				return nil, err
			}

			buildInfo := apiclient.CreateBuildInfo{
				DockerfileContent: img.Dockerfile(),
			}
			if len(contextHashes) > 0 {
				buildInfo.ContextHashes = contextHashes
			}
			createReq.SetBuildInfo(buildInfo)
		}
	}

	if resources != nil {
		if resources.CPU > 0 {
			createReq.SetCpu(int32(resources.CPU))
		}
		if resources.GPU > 0 {
			createReq.SetGpu(int32(resources.GPU))
		}
		if resources.Memory > 0 {
			createReq.SetMemory(int32(resources.Memory))
		}
		if resources.Disk > 0 {
			createReq.SetDisk(int32(resources.Disk))
		}
	}

	// Make API request using api-client-go
	authCtx := c.getAuthContext(ctx)
	sandboxResp, httpResp, err := c.apiClient.SandboxAPI.CreateSandbox(authCtx).CreateSandbox(*createReq).Execute()
	if err != nil {
		return nil, errors.ConvertAPIError(err, httpResp)
	}

	if sandboxResp.GetState() == apiclient.SANDBOXSTATE_ERROR || sandboxResp.GetState() == apiclient.SANDBOXSTATE_BUILD_FAILED {
		return nil, errors.NewDaytonaError("Sandbox failed to start", 0, nil)
	}

	toolboxClient, err := c.createToolboxClient(ctx, sandboxResp.GetId(), sandboxResp.GetTarget())
	if err != nil {
		return nil, err
	}

	autoArchiveInterval := 0
	if sandboxResp.AutoArchiveInterval != nil {
		autoArchiveInterval = int(*sandboxResp.AutoArchiveInterval)
	}

	autoDeleteInterval := 0
	if sandboxResp.AutoDeleteInterval != nil {
		autoDeleteInterval = int(*sandboxResp.AutoDeleteInterval)
	}

	sandbox := NewSandbox(c, toolboxClient, sandboxResp.GetId(), sandboxResp.GetName(), sandboxResp.GetState(), sandboxResp.GetTarget(), autoArchiveInterval, autoDeleteInterval, sandboxResp.GetNetworkBlockAll(), sandboxResp.NetworkAllowList)

	// Handle snapshot build logs
	if sandbox.State == apiclient.SANDBOXSTATE_PENDING_BUILD {
		if createOpts.LogChannel != nil {
			// Start log streaming in background
			go func() {
				defer close(createOpts.LogChannel)

				// Wait for sandbox to transition from PENDING_BUILD state
				ticker := time.NewTicker(1 * time.Second)
				defer ticker.Stop()

				for sandbox.State == apiclient.SANDBOXSTATE_PENDING_BUILD {
					select {
					case <-ctx.Done():
						return
					case <-ticker.C:
						// Refresh sandbox data
						if err := sandbox.RefreshData(ctx); err != nil {
							return
						}
					}
				}

				// Stream build logs after state transition
				if err := c.streamBuildLogsToChannel(ctx, sandbox.ID, createOpts.LogChannel); err != nil {
					// Optionally send error as final message
					select {
					case createOpts.LogChannel <- fmt.Sprintf("Error streaming logs: %v", err):
					case <-ctx.Done():
					}
				}
			}()
		}
	}

	// Wait for sandbox to start
	if createOpts.WaitForStart && sandbox.State != apiclient.SANDBOXSTATE_STARTED {
		if err := sandbox.WaitForStart(ctx, timeoutVal); err != nil {
			return nil, err
		}
	}

	return sandbox, nil
}

// Get retrieves an existing sandbox by its ID or name.
//
// The sandboxIDOrName parameter accepts either the sandbox's unique ID or its
// human-readable name. If a sandbox with the given identifier is not found,
// a [errors.DaytonaNotFoundError] is returned.
//
// Example:
//
//	sandbox, err := client.Get(ctx, "my-sandbox")
//	if err != nil {
//	    var notFound *errors.DaytonaNotFoundError
//	    if errors.As(err, &notFound) {
//	        log.Println("Sandbox not found")
//	    }
//	    return err
//	}
func (c *Client) Get(ctx context.Context, sandboxIDOrName string) (*Sandbox, error) {
	return withInstrumentation(ctx, c.Otel, "Client", "Get", func(ctx context.Context) (*Sandbox, error) {
		return c.doGet(ctx, sandboxIDOrName)
	})
}

func (c *Client) doGet(ctx context.Context, sandboxIDOrName string) (*Sandbox, error) {
	if sandboxIDOrName == "" {
		return nil, errors.NewDaytonaError("sandbox ID or name is required", 0, nil)
	}

	authCtx := c.getAuthContext(ctx)
	sandboxResp, httpResp, err := c.apiClient.SandboxAPI.GetSandbox(authCtx, sandboxIDOrName).Execute()
	if err != nil {
		return nil, errors.ConvertAPIError(err, httpResp)
	}

	toolboxClient, err := c.createToolboxClient(ctx, sandboxResp.GetId(), sandboxResp.GetTarget())
	if err != nil {
		return nil, err
	}

	autoArchiveInterval := 0
	if sandboxResp.AutoArchiveInterval != nil {
		autoArchiveInterval = int(*sandboxResp.AutoArchiveInterval)
	}

	autoDeleteInterval := 0
	if sandboxResp.AutoDeleteInterval != nil {
		autoDeleteInterval = int(*sandboxResp.AutoDeleteInterval)
	}

	// Capture sandbox state
	sandbox := NewSandbox(c,
		toolboxClient,
		sandboxResp.GetId(),
		sandboxResp.GetName(),
		sandboxResp.GetState(),
		sandboxResp.GetTarget(),
		autoArchiveInterval,
		autoDeleteInterval,
		sandboxResp.GetNetworkBlockAll(),
		sandboxResp.NetworkAllowList,
	)
	return sandbox, nil
}

// FindOne finds a single sandbox by ID/name or by matching labels.
//
// If sandboxIDOrName is provided and non-empty, FindOne delegates to [Client.Get].
// Otherwise, it searches for sandboxes matching the provided labels and returns
// the first match.
//
// This method is useful when you need to find a sandbox but may have either its
// identifier or its labels:
//
//	// Find by name
//	name := "my-sandbox"
//	sandbox, err := client.FindOne(ctx, &name, nil)
//
//	// Find by labels
//	sandbox, err := client.FindOne(ctx, nil, map[string]string{
//	    "environment": "production",
//	    "team":        "backend",
//	})
//
// Returns [errors.DaytonaNotFoundError] if no matching sandbox is found.
func (c *Client) FindOne(ctx context.Context, sandboxIDOrName *string, labels map[string]string) (*Sandbox, error) {
	return withInstrumentation(ctx, c.Otel, "Client", "FindOne", func(ctx context.Context) (*Sandbox, error) {
		return c.doFindOne(ctx, sandboxIDOrName, labels)
	})
}

func (c *Client) doFindOne(ctx context.Context, sandboxIDOrName *string, labels map[string]string) (*Sandbox, error) {
	if sandboxIDOrName != nil && *sandboxIDOrName != "" {
		return c.Get(ctx, *sandboxIDOrName)
	}

	pages := 1
	limit := 1
	result, err := c.List(ctx, labels, &pages, &limit)
	if err != nil {
		return nil, err
	}

	if len(result.Items) == 0 {
		labelsJSON, _ := json.Marshal(labels)
		return nil, errors.NewDaytonaNotFoundError(fmt.Sprintf("No sandbox found with labels %s", labelsJSON), nil)
	}

	return result.Items[0], nil
}

// List retrieves sandboxes with optional label filtering and pagination.
//
// Parameters:
//   - labels: Optional map of labels to filter sandboxes. Pass nil for no filtering.
//   - page: Optional page number (1-indexed). Pass nil for the first page.
//   - limit: Optional number of results per page. Pass nil for the default limit.
//
// Example:
//
//	// List all sandboxes
//	result, err := client.List(ctx, nil, nil, nil)
//
//	// List sandboxes with pagination
//	page, limit := 1, 10
//	result, err := client.List(ctx, nil, &page, &limit)
//
//	// Filter by labels
//	result, err := client.List(ctx, map[string]string{"env": "dev"}, nil, nil)
//
//	// Iterate through results
//	for _, sandbox := range result.Items {
//	    fmt.Printf("Sandbox: %s (state: %s)\n", sandbox.Name, sandbox.State)
//	}
//
// Returns a [PaginatedSandboxes] containing the matching sandboxes and pagination metadata.
func (c *Client) List(ctx context.Context, labels map[string]string, page *int, limit *int) (*PaginatedSandboxes, error) {
	return withInstrumentation(ctx, c.Otel, "Client", "List", func(ctx context.Context) (*PaginatedSandboxes, error) {
		return c.doList(ctx, labels, page, limit)
	})
}

func (c *Client) doList(ctx context.Context, labels map[string]string, page *int, limit *int) (*PaginatedSandboxes, error) {
	if page != nil && *page < 1 {
		return nil, errors.NewDaytonaError("page must be a positive integer", 0, nil)
	}
	if limit != nil && *limit < 1 {
		return nil, errors.NewDaytonaError("limit must be a positive integer", 0, nil)
	}

	authCtx := c.getAuthContext(ctx)
	request := c.apiClient.SandboxAPI.ListSandboxesPaginated(authCtx)

	// Add optional parameters
	if labels != nil {
		labelsJSON, _ := json.Marshal(labels)
		request = request.Labels(string(labelsJSON))
	}
	if page != nil {
		request = request.Page(float32(*page))
	}
	if limit != nil {
		request = request.Limit(float32(*limit))
	}

	result, httpResp, err := request.Execute()
	if err != nil {
		return nil, errors.ConvertAPIError(err, httpResp)
	}

	items := result.GetItems()
	sandboxes := make([]*Sandbox, len(items))
	for i := range items {
		toolboxClient, err := c.createToolboxClient(ctx, items[i].GetId(), items[i].GetTarget())
		if err != nil {
			return nil, err
		}

		autoArchiveInterval := 0
		if items[i].AutoArchiveInterval != nil {
			autoArchiveInterval = int(*items[i].AutoArchiveInterval)
		}

		autoDeleteInterval := 0
		if items[i].AutoDeleteInterval != nil {
			autoDeleteInterval = int(*items[i].AutoDeleteInterval)
		}

		sandboxes[i] = NewSandbox(c,
			toolboxClient,
			items[i].GetId(),
			items[i].GetName(),
			items[i].GetState(),
			items[i].GetTarget(),
			autoArchiveInterval,
			autoDeleteInterval,
			items[i].GetNetworkBlockAll(),
			items[i].NetworkAllowList,
		)
	}

	return &PaginatedSandboxes{
		Items:      sandboxes,
		Total:      int(result.GetTotal()),
		Page:       int(result.GetPage()),
		TotalPages: int(result.GetTotalPages()),
	}, nil
}

// streamBuildLogsToChannel streams build logs for a sandbox to a channel
func (c *Client) streamBuildLogsToChannel(ctx context.Context, sandboxID string, logChan chan<- string) error {
	// Build the URL for the sandbox build logs endpoint
	cfg := c.apiClient.GetConfig()
	baseURL := cfg.Servers[0].URL
	url := fmt.Sprintf("%s/sandbox/%s/build-logs?follow=true", baseURL, sandboxID)

	// Create HTTP request
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Add authentication and headers
	token := c.apiKey
	if token == "" {
		token = c.jwtToken
	}
	httpReq.Header.Set("Authorization", "Bearer "+token)

	// Add default headers from API client config
	for key, value := range cfg.DefaultHeader {
		httpReq.Header.Set(key, value)
	}

	// Add organization header if using JWT
	if c.jwtToken != "" && c.organizationID != "" {
		httpReq.Header.Set("X-Daytona-Organization-ID", c.organizationID)
	}

	// Execute the HTTP request
	resp, err := cfg.HTTPClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	// Stream the response line by line
	reader := bufio.NewReader(resp.Body)
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			line, err := reader.ReadString('\n')
			if err != nil {
				if err == io.EOF {
					// End of stream
					if line != "" {
						select {
						case logChan <- strings.TrimRight(line, "\n\r"):
						case <-ctx.Done():
						}
					}
					return nil
				}
				return fmt.Errorf("failed to read log line: %w", err)
			}

			// Send log line to channel (remove trailing newline)
			if line != "" {
				select {
				case logChan <- strings.TrimRight(line, "\n\r"):
				case <-ctx.Done():
					return ctx.Err()
				}
			}
		}
	}
}

// getProxyToolboxURL gets the proxy toolbox URL for a specific region.
// The URL is cached per region to avoid redundant API calls.
func (c *Client) getProxyToolboxURL(ctx context.Context, sandboxID string, regionID string) (string, error) {
	// Check cache first
	if cachedURL, exists := c.toolboxProxyCache.Get(regionID); exists {
		return cachedURL, nil
	}

	// Fetch from API using the sandbox-specific endpoint
	authCtx := c.getAuthContext(ctx)
	result, httpResp, err := c.apiClient.SandboxAPI.GetToolboxProxyUrl(authCtx, sandboxID).Execute()
	if err != nil {
		return "", c.handleAPIError(err, httpResp)
	}

	proxyURL := result.GetUrl()

	// Cache the result by region
	c.toolboxProxyCache.Set(regionID, proxyURL)

	return proxyURL, nil
}

// getPushAccessCredentials gets object storage push access credentials from the API
func (c *Client) getPushAccessCredentials(ctx context.Context) (*PushAccessCredentials, error) {
	creds, httpResp, err := c.apiClient.ObjectStorageAPI.GetPushAccess(c.getAuthContext(ctx)).Execute()
	if err != nil {
		return nil, c.handleAPIError(err, httpResp)
	}

	// Map API response to internal structure
	result := &PushAccessCredentials{
		StorageURL:     creds.GetStorageUrl(),
		AccessKey:      creds.GetAccessKey(),
		Secret:         creds.GetSecret(),
		SessionToken:   creds.GetSessionToken(),
		Bucket:         creds.GetBucket(),
		OrganizationID: creds.GetOrganizationId(),
	}

	return result, nil
}

// processImageContext processes image contexts and uploads them to object storage
func (c *Client) processImageContext(ctx context.Context, image *DockerImage) ([]string, error) {
	contexts := image.Contexts()
	if len(contexts) == 0 {
		return []string{}, nil
	}

	// Get push access credentials
	creds, err := c.getPushAccessCredentials(ctx)
	if err != nil {
		return nil, err
	}

	// Create object storage client
	objStorage := NewObjectStorage(objectStorageConfig{
		EndpointURL:     creds.StorageURL,
		AccessKeyID:     creds.AccessKey,
		SecretAccessKey: creds.Secret,
		SessionToken:    &creds.SessionToken,
		BucketName:      creds.Bucket,
	})

	// Upload each context
	contextHashes := make([]string, 0, len(contexts))
	for _, context := range contexts {
		hash, err := objStorage.Upload(ctx, context.SourcePath, creds.OrganizationID, context.ArchivePath)
		if err != nil {
			return nil, err
		}
		contextHashes = append(contextHashes, hash)
	}

	return contextHashes, nil
}
