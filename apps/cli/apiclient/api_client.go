// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package apiclient

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"sync"

	"github.com/daytonaio/daytona/cli/auth"
	"github.com/daytonaio/daytona/cli/config"
	"github.com/daytonaio/daytona/cli/internal"
	apiclient "github.com/daytonaio/daytona/libs/api-client-go"

	log "github.com/sirupsen/logrus"
)

type versionCheckTransport struct {
	transport http.RoundTripper
}

var versionMismatchWarningOnce sync.Once

func (t *versionCheckTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	resp, err := t.transport.RoundTrip(req)
	if resp != nil {
		// Check version mismatch on all responses, not just errors
		checkVersionsMismatch(resp)
	}
	return resp, err
}

var apiClient *apiclient.APIClient

const DaytonaSourceHeader = "X-Daytona-Source"
const API_VERSION_HEADER = "X-Daytona-Api-Version"

func checkVersionsMismatch(res *http.Response) {
	// If the CLI is running in a structured output mode (e.g. json/yaml),
	// avoid printing human-readable warnings that could break consumers.
	if internal.SuppressVersionMismatchWarning {
		return
	}

	serverVersion := res.Header.Get(API_VERSION_HEADER)
	if serverVersion == "" {
		return
	}

	// Trim "v" prefix from both versions for comparison
	cliVersion := strings.TrimPrefix(internal.Version, "v")
	apiVersion := strings.TrimPrefix(serverVersion, "v")

	if cliVersion == "0.0.0-dev" || cliVersion == apiVersion {
		return
	}

	if compareVersions(cliVersion, apiVersion) >= 0 {
		return
	}

	versionMismatchWarningOnce.Do(func() {
		log.Warn(fmt.Sprintf("Version mismatch: Daytona CLI is on v%s and API is on v%s.\nMake sure the versions are aligned using 'brew upgrade daytonaio/cli/daytona' or by downloading the latest version from https://github.com/daytonaio/daytona/releases.", cliVersion, apiVersion))
	})
}

// compareVersions compares two semver strings
// Returns: -1 if v1 < v2, 0 if v1 == v2, 1 if v1 > v2
func compareVersions(v1, v2 string) int {
	parts1 := strings.Split(v1, ".")
	parts2 := strings.Split(v2, ".")

	maxLen := len(parts1)
	if len(parts2) > maxLen {
		maxLen = len(parts2)
	}

	for i := 0; i < maxLen; i++ {
		var n1, n2 int
		if i < len(parts1) {
			fmt.Sscanf(parts1[i], "%d", &n1)
		}
		if i < len(parts2) {
			fmt.Sscanf(parts2[i], "%d", &n2)
		}

		if n1 < n2 {
			return -1
		}
		if n1 > n2 {
			return 1
		}
	}

	return 0
}

func GetApiClient(profile *config.Profile, defaultHeaders map[string]string) (*apiclient.APIClient, error) {
	c, err := config.GetConfig()
	if err != nil {
		return nil, err
	}

	var activeProfile config.Profile
	if profile == nil {
		var err error
		activeProfile, err = c.GetActiveProfile()
		if err != nil {
			return nil, err
		}
	} else {
		activeProfile = *profile
	}

	if apiClient != nil && activeProfile.Api.Key == nil {
		err := auth.RefreshTokenIfNeeded(context.Background())
		if err != nil {
			return nil, err
		}

		return apiClient, nil
	}

	var newApiClient *apiclient.APIClient

	serverUrl := activeProfile.Api.Url

	clientConfig := apiclient.NewConfiguration()
	clientConfig.Servers = apiclient.ServerConfigurations{
		{
			URL: serverUrl,
		},
	}

	if activeProfile.Api.Key != nil {
		clientConfig.AddDefaultHeader("Authorization", "Bearer "+*activeProfile.Api.Key)
	} else if activeProfile.Api.Token != nil {
		clientConfig.AddDefaultHeader("Authorization", "Bearer "+activeProfile.Api.Token.AccessToken)

		if activeProfile.ActiveOrganizationId != nil {
			clientConfig.AddDefaultHeader("X-Daytona-Organization-ID", *activeProfile.ActiveOrganizationId)
		}
	}

	clientConfig.AddDefaultHeader(DaytonaSourceHeader, "cli")

	for headerKey, headerValue := range defaultHeaders {
		clientConfig.AddDefaultHeader(headerKey, headerValue)
	}

	newApiClient = apiclient.NewAPIClient(clientConfig)

	newApiClient.GetConfig().HTTPClient = &http.Client{
		Transport: &versionCheckTransport{
			transport: http.DefaultTransport,
		},
	}

	if apiClient != nil && activeProfile.Api.Key == nil {
		err = auth.RefreshTokenIfNeeded(context.Background())
		if err != nil {
			return nil, err
		}
	}

	apiClient = newApiClient
	return apiClient, nil
}
