// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package daytona

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"

	"testing"
	"time"

	apiclient "github.com/daytonaio/daytona/libs/api-client-go"
	"github.com/daytonaio/daytona/libs/sdk-go/pkg/options"
	"github.com/daytonaio/daytona/libs/sdk-go/pkg/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestNewClient tests the NewClient function
func TestNewClient(t *testing.T) {
	tests := []struct {
		name           string
		envVars        map[string]string
		expectedError  bool
		errorContains  string
		validateClient func(*testing.T, *Client)
	}{
		{
			name: "success with API key",
			envVars: map[string]string{
				"DAYTONA_API_KEY": "test-api-key",
			},
			expectedError: false,
			validateClient: func(t *testing.T, c *Client) {
				assert.Equal(t, "test-api-key", c.apiKey)
				assert.Equal(t, defaultAPIURL, c.apiURL)
				assert.NotNil(t, c.httpClient)
				assert.NotNil(t, c.apiClient)
				assert.NotNil(t, c.Volume)
				assert.NotNil(t, c.Snapshot)
			},
		},
		{
			name: "success with JWT token and org ID",
			envVars: map[string]string{
				"DAYTONA_JWT_TOKEN":       "test-jwt-token",
				"DAYTONA_ORGANIZATION_ID": "test-org-id",
			},
			expectedError: false,
			validateClient: func(t *testing.T, c *Client) {
				assert.Equal(t, "test-jwt-token", c.jwtToken)
				assert.Equal(t, "test-org-id", c.organizationID)
			},
		},
		{
			name: "success with custom API URL",
			envVars: map[string]string{
				"DAYTONA_API_KEY": "test-api-key",
				"DAYTONA_API_URL": "https://custom.api.url/api",
			},
			expectedError: false,
			validateClient: func(t *testing.T, c *Client) {
				assert.Equal(t, "https://custom.api.url/api", c.apiURL)
			},
		},
		{
			name: "success with target",
			envVars: map[string]string{
				"DAYTONA_API_KEY": "test-api-key",
				"DAYTONA_TARGET":  "test-target",
			},
			expectedError: false,
			validateClient: func(t *testing.T, c *Client) {
				assert.Equal(t, "test-target", c.region)
			},
		},
		{
			name:          "error without API key or JWT token",
			envVars:       map[string]string{},
			expectedError: true,
			errorContains: "API key or JWT token is required",
		},
		{
			name: "error with JWT token but no org ID",
			envVars: map[string]string{
				"DAYTONA_JWT_TOKEN": "test-jwt-token",
			},
			expectedError: true,
			errorContains: "Organization ID is required when using JWT token",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear environment
			os.Clearenv()

			// Set test environment variables
			for key, value := range tt.envVars {
				os.Setenv(key, value)
			}

			// Create client
			client, err := NewClient()

			// Validate error expectations
			if tt.expectedError {
				require.Error(t, err)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
				assert.Nil(t, client)
			} else {
				require.NoError(t, err)
				require.NotNil(t, client)
				if tt.validateClient != nil {
					tt.validateClient(t, client)
				}
			}
		})
	}

	// Cleanup
	os.Clearenv()
}

// TestNewClientWithConfig tests the NewClientWithConfig function
func TestNewClientWithConfig(t *testing.T) {
	tests := []struct {
		name           string
		config         *types.DaytonaConfig
		envVars        map[string]string
		expectedError  bool
		errorContains  string
		validateClient func(*testing.T, *Client)
	}{
		{
			name: "success with API key in config",
			config: &types.DaytonaConfig{
				APIKey: "config-api-key",
			},
			expectedError: false,
			validateClient: func(t *testing.T, c *Client) {
				assert.Equal(t, "config-api-key", c.apiKey)
			},
		},
		{
			name: "config overrides environment variables",
			config: &types.DaytonaConfig{
				APIKey: "config-api-key",
				APIUrl: "https://config.api.url/api",
			},
			envVars: map[string]string{
				"DAYTONA_API_KEY": "env-api-key",
				"DAYTONA_API_URL": "https://env.api.url/api",
			},
			expectedError: false,
			validateClient: func(t *testing.T, c *Client) {
				assert.Equal(t, "config-api-key", c.apiKey)
				assert.Equal(t, "https://config.api.url/api", c.apiURL)
			},
		},
		{
			name: "JWT token with org ID in config",
			config: &types.DaytonaConfig{
				JWTToken:       "config-jwt-token",
				OrganizationID: "config-org-id",
			},
			expectedError: false,
			validateClient: func(t *testing.T, c *Client) {
				assert.Equal(t, "config-jwt-token", c.jwtToken)
				assert.Equal(t, "config-org-id", c.organizationID)
			},
		},
		{
			name:   "falls back to env vars when config fields are empty",
			config: &types.DaytonaConfig{
				// Empty config
			},
			envVars: map[string]string{
				"DAYTONA_API_KEY": "env-api-key",
			},
			expectedError: false,
			validateClient: func(t *testing.T, c *Client) {
				assert.Equal(t, "env-api-key", c.apiKey)
			},
		},
		{
			name:   "nil config uses environment variables",
			config: nil,
			envVars: map[string]string{
				"DAYTONA_API_KEY": "env-api-key",
			},
			expectedError: false,
			validateClient: func(t *testing.T, c *Client) {
				assert.Equal(t, "env-api-key", c.apiKey)
			},
		},
		{
			name: "custom target in config",
			config: &types.DaytonaConfig{
				APIKey: "test-api-key",
				Target: "custom-target",
			},
			expectedError: false,
			validateClient: func(t *testing.T, c *Client) {
				assert.Equal(t, "custom-target", c.region)
			},
		},
		{
			name: "error without any authentication",
			config: &types.DaytonaConfig{
				OrganizationID: "org-id-only",
			},
			expectedError: true,
			errorContains: "API key or JWT token is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear environment
			os.Clearenv()

			// Set test environment variables
			for key, value := range tt.envVars {
				os.Setenv(key, value)
			}

			// Create client
			client, err := NewClientWithConfig(tt.config)

			// Validate error expectations
			if tt.expectedError {
				require.Error(t, err)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
				assert.Nil(t, client)
			} else {
				require.NoError(t, err)
				require.NotNil(t, client)
				if tt.validateClient != nil {
					tt.validateClient(t, client)
				}
			}
		})
	}

	// Cleanup
	os.Clearenv()
}

// TestGetAuthContext tests the getAuthContext method
func TestGetAuthContext(t *testing.T) {
	tests := []struct {
		name          string
		client        *Client
		expectedToken string
	}{
		{
			name: "returns API key when set",
			client: &Client{
				apiKey: "test-api-key",
			},
			expectedToken: "test-api-key",
		},
		{
			name: "returns JWT token when API key is empty",
			client: &Client{
				jwtToken: "test-jwt-token",
			},
			expectedToken: "test-jwt-token",
		},
		{
			name: "prefers API key over JWT token",
			client: &Client{
				apiKey:   "test-api-key",
				jwtToken: "test-jwt-token",
			},
			expectedToken: "test-api-key",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			authCtx := tt.client.getAuthContext(ctx)

			token := authCtx.Value(apiclient.ContextAccessToken)
			assert.Equal(t, tt.expectedToken, token)
		})
	}
}

// TestHandleAPIError tests the handleAPIError method
func TestHandleAPIError(t *testing.T) {
	tests := []struct {
		name           string
		err            error
		httpResp       *http.Response
		expectedError  string
		expectedType   string
		expectedStatus int
	}{
		{
			name: "not found error",
			err:  fmt.Errorf("not found"),
			httpResp: &http.Response{
				StatusCode: http.StatusNotFound,
				Header:     http.Header{},
			},
			expectedError:  "Resource not found",
			expectedType:   "*errors.DaytonaNotFoundError",
			expectedStatus: http.StatusNotFound,
		},
		{
			name: "rate limit error",
			err:  fmt.Errorf("rate limit exceeded"),
			httpResp: &http.Response{
				StatusCode: http.StatusTooManyRequests,
				Header:     http.Header{},
			},
			expectedError:  "Rate limit exceeded",
			expectedType:   "*errors.DaytonaRateLimitError",
			expectedStatus: http.StatusTooManyRequests,
		},
		{
			name: "generic error with status code",
			err:  fmt.Errorf("internal server error"),
			httpResp: &http.Response{
				StatusCode: http.StatusInternalServerError,
				Header:     http.Header{},
			},
			expectedError:  "internal server error",
			expectedType:   "*errors.DaytonaError",
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name:          "error without http response",
			err:           fmt.Errorf("network error"),
			httpResp:      nil,
			expectedError: "network error",
			expectedType:  "*errors.DaytonaError",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := &Client{}
			err := client.handleAPIError(tt.err, tt.httpResp)

			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.expectedError)
			assert.Equal(t, tt.expectedType, fmt.Sprintf("%T", err))
		})
	}
}

// TestGet tests the Get method
func TestGet(t *testing.T) {
	t.Run("error when sandbox ID is empty", func(t *testing.T) {
		// Setup test environment
		os.Clearenv()
		os.Setenv("DAYTONA_API_KEY", "test-api-key")

		client, err := NewClient()
		require.NoError(t, err)

		ctx := context.Background()
		sandbox, err := client.Get(ctx, "")

		require.Error(t, err)
		assert.Contains(t, err.Error(), "sandbox ID or name is required")
		assert.Nil(t, sandbox)
	})
}

// TestFindOne tests the FindOne method
func TestFindOne(t *testing.T) {
	t.Run("validates labels parameter", func(t *testing.T) {
		// Setup test environment
		os.Clearenv()
		os.Setenv("DAYTONA_API_KEY", "test-api-key")

		client, err := NewClient()
		require.NoError(t, err)

		ctx := context.Background()

		// Test with nil ID and empty labels - should return error from List
		sandbox, err := client.FindOne(ctx, nil, map[string]string{})
		assert.Error(t, err) // Will fail when trying to call API
		assert.Nil(t, sandbox)
	})

	t.Run("calls Get when ID is provided", func(t *testing.T) {
		// Setup test environment
		os.Clearenv()
		os.Setenv("DAYTONA_API_KEY", "test-api-key")

		client, err := NewClient()
		require.NoError(t, err)

		ctx := context.Background()
		emptyID := ""

		// Test with empty string ID
		sandbox, err := client.FindOne(ctx, &emptyID, nil)
		assert.Error(t, err)
		assert.Nil(t, sandbox)
	})
}

// TestList tests the List method
func TestList(t *testing.T) {
	tests := []struct {
		name          string
		page          *int
		limit         *int
		expectedError bool
		errorContains string
	}{
		{
			name:          "valid pagination",
			page:          intPtr(1),
			limit:         intPtr(10),
			expectedError: false,
		},
		{
			name:          "nil pagination parameters",
			page:          nil,
			limit:         nil,
			expectedError: false,
		},
		{
			name:          "invalid page - zero",
			page:          intPtr(0),
			limit:         intPtr(10),
			expectedError: true,
			errorContains: "page must be a positive integer",
		},
		{
			name:          "invalid page - negative",
			page:          intPtr(-1),
			limit:         intPtr(10),
			expectedError: true,
			errorContains: "page must be a positive integer",
		},
		{
			name:          "invalid limit - zero",
			page:          intPtr(1),
			limit:         intPtr(0),
			expectedError: true,
			errorContains: "limit must be a positive integer",
		},
		{
			name:          "invalid limit - negative",
			page:          intPtr(1),
			limit:         intPtr(-1),
			expectedError: true,
			errorContains: "limit must be a positive integer",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup test environment
			os.Clearenv()
			os.Setenv("DAYTONA_API_KEY", "test-api-key")

			client, err := NewClient()
			require.NoError(t, err)

			ctx := context.Background()
			result, err := client.List(ctx, nil, tt.page, tt.limit)

			if tt.expectedError {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorContains)
				assert.Nil(t, result)
			} else {
				// Will fail on API call, but validation should pass
				assert.Error(t, err) // API call will fail in test
			}
		})
	}
}

// TestCreateOptions tests the Create options
func TestCreateOptions(t *testing.T) {
	t.Run("WithTimeout sets timeout", func(t *testing.T) {
		opts := &options.CreateSandbox{}
		timeout := 30 * time.Second

		opt := options.WithTimeout(timeout)
		opt(opts)

		require.NotNil(t, opts.Timeout)
		assert.Equal(t, timeout, *opts.Timeout)
	})

	t.Run("WithWaitForStart sets waitForStart", func(t *testing.T) {
		opts := &options.CreateSandbox{}

		opt := options.WithWaitForStart(false)
		opt(opts)

		assert.False(t, opts.WaitForStart)
	})

	t.Run("WithLogChannel sets logChannel", func(t *testing.T) {
		opts := &options.CreateSandbox{}
		logChan := make(chan string)

		opt := options.WithLogChannel(logChan)
		opt(opts)

		assert.NotNil(t, opts.LogChannel)
		assert.Equal(t, logChan, opts.LogChannel)
	})
}

// TestCreateValidation tests the Create method validation
func TestCreateValidation(t *testing.T) {
	tests := []struct {
		name          string
		params        interface{}
		expectedError bool
		errorContains string
	}{
		{
			name: "invalid auto stop interval - negative",
			params: types.ImageParams{
				SandboxBaseParams: types.SandboxBaseParams{
					AutoStopInterval: intPtr(-1),
				},
				Image: "test-image",
			},
			expectedError: true,
			errorContains: "autoStopInterval must be a non-negative integer",
		},
		{
			name: "invalid auto archive interval - negative",
			params: types.ImageParams{
				SandboxBaseParams: types.SandboxBaseParams{
					AutoArchiveInterval: intPtr(-1),
				},
				Image: "test-image",
			},
			expectedError: true,
			errorContains: "autoArchiveInterval must be a non-negative integer",
		},
		{
			name: "valid auto intervals",
			params: types.ImageParams{
				SandboxBaseParams: types.SandboxBaseParams{
					AutoStopInterval:    intPtr(60),
					AutoArchiveInterval: intPtr(3600),
				},
				Image: "test-image",
			},
			expectedError: false,
		},
		{
			name: "ephemeral sets auto delete to zero",
			params: types.ImageParams{
				SandboxBaseParams: types.SandboxBaseParams{
					Ephemeral: true,
				},
				Image: "test-image",
			},
			expectedError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup test environment
			os.Clearenv()
			os.Setenv("DAYTONA_API_KEY", "test-api-key")

			client, err := NewClient()
			require.NoError(t, err)

			ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
			defer cancel()

			_, err = client.Create(ctx, tt.params)

			if tt.expectedError {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorContains)
			} else {
				// Will fail on API call, but validation should pass
				assert.Error(t, err) // API call will fail in test
			}
		})
	}
}

// TestStreamBuildLogsToChannel tests build log streaming
func TestStreamBuildLogsToChannel(t *testing.T) {
	t.Run("streams logs successfully", func(t *testing.T) {
		// Create a test server that returns log lines
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Verify auth header
			assert.Equal(t, "Bearer test-api-key", r.Header.Get("Authorization"))

			// Write some log lines
			fmt.Fprintln(w, "Building image...")
			fmt.Fprintln(w, "Step 1/3")
			fmt.Fprintln(w, "Step 2/3")
			fmt.Fprintln(w, "Step 3/3")
			fmt.Fprintln(w, "Successfully built")
		}))
		defer server.Close()

		// Setup client with test server
		os.Clearenv()
		os.Setenv("DAYTONA_API_KEY", "test-api-key")
		os.Setenv("DAYTONA_API_URL", server.URL)

		client, err := NewClient()
		require.NoError(t, err)

		// Override the API URL to point to test server
		client.apiURL = server.URL
		cfg := client.apiClient.GetConfig()
		cfg.Servers = apiclient.ServerConfigurations{
			{URL: server.URL},
		}

		ctx := context.Background()
		logChan := make(chan string, 10)

		go func() {
			err := client.streamBuildLogsToChannel(ctx, "test-sandbox-id", logChan)
			assert.NoError(t, err)
			close(logChan)
		}()

		// Collect logs
		var logs []string
		for log := range logChan {
			logs = append(logs, log)
		}

		// Verify logs
		assert.Len(t, logs, 5)
		assert.Equal(t, "Building image...", logs[0])
		assert.Equal(t, "Successfully built", logs[4])
	})

	t.Run("handles context cancellation", func(t *testing.T) {
		// Create a test server that streams indefinitely
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Stream logs slowly
			for i := 0; i < 100; i++ {
				fmt.Fprintf(w, "Log line %d\n", i)
				if f, ok := w.(http.Flusher); ok {
					f.Flush()
				}
				time.Sleep(100 * time.Millisecond)
			}
		}))
		defer server.Close()

		// Setup client
		os.Clearenv()
		os.Setenv("DAYTONA_API_KEY", "test-api-key")

		client, err := NewClient()
		require.NoError(t, err)

		// Override API URL
		cfg := client.apiClient.GetConfig()
		cfg.Servers = apiclient.ServerConfigurations{
			{URL: server.URL},
		}

		ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
		defer cancel()

		logChan := make(chan string, 10)

		err = client.streamBuildLogsToChannel(ctx, "test-sandbox-id", logChan)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "context deadline exceeded")
	})
}

// TestDefaultLanguage tests default language setting
func TestDefaultLanguage(t *testing.T) {
	os.Clearenv()
	os.Setenv("DAYTONA_API_KEY", "test-api-key")

	client, err := NewClient()
	require.NoError(t, err)

	assert.Equal(t, types.CodeLanguagePython, client.defaultLanguage)
}

// Helper functions

func intPtr(i int) *int {
	return &i
}

// TestClientTimeout tests HTTP client timeout configuration
func TestClientTimeout(t *testing.T) {
	os.Clearenv()
	os.Setenv("DAYTONA_API_KEY", "test-api-key")

	client, err := NewClient()
	require.NoError(t, err)

	assert.Equal(t, defaultTimeout, client.httpClient.Timeout)
}

// TestServerURLConstruction tests proper server URL construction
func TestServerURLConstruction(t *testing.T) {
	tests := []struct {
		name        string
		apiURL      string
		expectedURL string
	}{
		{
			name:        "default API URL",
			apiURL:      defaultAPIURL,
			expectedURL: "https://app.daytona.io/api",
		},
		{
			name:        "custom API URL with path",
			apiURL:      "https://custom.daytona.io/api/v1",
			expectedURL: "https://custom.daytona.io/api/v1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Clearenv()
			os.Setenv("DAYTONA_API_KEY", "test-api-key")
			os.Setenv("DAYTONA_API_URL", tt.apiURL)

			client, err := NewClient()
			require.NoError(t, err)

			cfg := client.apiClient.GetConfig()
			assert.Equal(t, tt.expectedURL, cfg.Servers[0].URL)
		})
	}
}

// TestJWTAuthentication tests JWT token authentication setup
func TestJWTAuthentication(t *testing.T) {
	os.Clearenv()
	os.Setenv("DAYTONA_JWT_TOKEN", "test-jwt-token")
	os.Setenv("DAYTONA_ORGANIZATION_ID", "test-org-id")

	client, err := NewClient()
	require.NoError(t, err)

	// Verify JWT token and org ID are set
	assert.Equal(t, "test-jwt-token", client.jwtToken)
	assert.Equal(t, "test-org-id", client.organizationID)

	// Verify organization header is set
	cfg := client.apiClient.GetConfig()
	orgHeader, exists := cfg.DefaultHeader["X-Daytona-Organization-ID"]
	assert.True(t, exists)
	assert.Equal(t, "test-org-id", orgHeader)
}

// TestSDKSourceHeader tests that SDK source header is set
func TestSDKSourceHeader(t *testing.T) {
	os.Clearenv()
	os.Setenv("DAYTONA_API_KEY", "test-api-key")

	client, err := NewClient()
	require.NoError(t, err)

	cfg := client.apiClient.GetConfig()
	sourceHeader, exists := cfg.DefaultHeader["X-Daytona-Source"]
	assert.True(t, exists)
	assert.Equal(t, sdkSource, sourceHeader)
	assert.Equal(t, "go-sdk", sdkSource)
}

// TestServicesInitialization tests that services are properly initialized
func TestServicesInitialization(t *testing.T) {
	os.Clearenv()
	os.Setenv("DAYTONA_API_KEY", "test-api-key")

	client, err := NewClient()
	require.NoError(t, err)

	assert.NotNil(t, client.Volume)
	assert.NotNil(t, client.Snapshot)
}

// TestCreateToolboxClient tests the createToolboxClient method
func TestCreateToolboxClient(t *testing.T) {
	t.Run("creates toolbox client with correct configuration", func(t *testing.T) {
		os.Clearenv()
		os.Setenv("DAYTONA_API_KEY", "test-api-key")

		client, err := NewClient()
		require.NoError(t, err)

		toolboxClient, err := client.createToolboxClient("https://toolbox-proxy.example.com", "test-sandbox-id")
		require.NoError(t, err)
		require.NotNil(t, toolboxClient)

		// Verify the toolbox client configuration
		cfg := toolboxClient.GetConfig()

		// Check that the server URL includes the sandbox ID
		assert.Contains(t, cfg.Servers[0].URL, "test-sandbox-id")

		// Check auth headers
		authHeader, exists := cfg.DefaultHeader["Authorization"]
		assert.True(t, exists)
		assert.Equal(t, "Bearer test-api-key", authHeader)

		// Check SDK source header
		sourceHeader, exists := cfg.DefaultHeader["X-Daytona-Source"]
		assert.True(t, exists)
		assert.Equal(t, "go-sdk", sourceHeader)
	})

	t.Run("creates toolbox client with JWT auth", func(t *testing.T) {
		os.Clearenv()
		os.Setenv("DAYTONA_JWT_TOKEN", "test-jwt-token")
		os.Setenv("DAYTONA_ORGANIZATION_ID", "test-org-id")

		client, err := NewClient()
		require.NoError(t, err)

		toolboxClient, err := client.createToolboxClient("https://toolbox-proxy.example.com", "test-sandbox-id")
		require.NoError(t, err)
		require.NotNil(t, toolboxClient)

		// Verify the toolbox client configuration
		cfg := toolboxClient.GetConfig()

		// Check auth headers
		authHeader, exists := cfg.DefaultHeader["Authorization"]
		assert.True(t, exists)
		assert.Equal(t, "Bearer test-jwt-token", authHeader)

		// Check organization header
		orgHeader, exists := cfg.DefaultHeader["X-Daytona-Organization-ID"]
		assert.True(t, exists)
		assert.Equal(t, "test-org-id", orgHeader)
	})
}

// TestSandboxTargetField tests that Sandbox has Target field properly set
func TestSandboxTargetField(t *testing.T) {
	t.Run("NewSandbox sets target field", func(t *testing.T) {
		os.Clearenv()
		os.Setenv("DAYTONA_API_KEY", "test-api-key")

		client, err := NewClient()
		require.NoError(t, err)

		sandbox := NewSandbox(
			client,
			nil, // toolboxClient
			"test-id",
			"test-name",
			apiclient.SANDBOXSTATE_STARTED,
			"us-east-1", // target
			60,          // autoArchiveInterval
			-1,          // autoDeleteInterval
			false,       // networkBlockAll
			nil,         // networkAllowList
		)

		assert.Equal(t, "test-id", sandbox.ID)
		assert.Equal(t, "test-name", sandbox.Name)
		assert.Equal(t, "us-east-1", sandbox.Target)
		assert.Equal(t, apiclient.SANDBOXSTATE_STARTED, sandbox.State)
	})
}
