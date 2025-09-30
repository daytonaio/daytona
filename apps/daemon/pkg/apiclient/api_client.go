// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package apiclient

import (
	"fmt"
	"net/http"
	"sync"

	apiclient "github.com/daytonaio/apiclient"
)

type ApiAccess struct {
	ApiUrl string `json:"apiUrl"`
	Token  string `json:"token"`
}

var (
	apiClient     *apiclient.APIClient
	apiAccess     *ApiAccess
	apiClientLock sync.RWMutex
)

// SetApiAccess updates the API access information and reinitializes the client
func SetApiAccess(access *ApiAccess) error {
	if access == nil {
		return fmt.Errorf("api access cannot be nil")
	}

	apiClientLock.Lock()
	defer apiClientLock.Unlock()

	apiAccess = access
	apiClient = nil // Force recreation on next GetApiClient call
	return nil
}

// GetApiClient returns a singleton instance of the API client
func GetApiClient() (*apiclient.APIClient, error) {
	apiClientLock.RLock()
	if apiClient != nil {
		defer apiClientLock.RUnlock()
		return apiClient, nil
	}
	apiClientLock.RUnlock()

	apiClientLock.Lock()
	defer apiClientLock.Unlock()

	// Double-check after acquiring write lock
	if apiClient != nil {
		return apiClient, nil
	}

	if apiAccess == nil {
		return nil, fmt.Errorf("api access not configured")
	}

	clientConfig := apiclient.NewConfiguration()
	clientConfig.Servers = apiclient.ServerConfigurations{
		{
			URL: apiAccess.ApiUrl,
		},
	}

	clientConfig.AddDefaultHeader("Authorization", "Bearer "+apiAccess.Token)
	clientConfig.AddDefaultHeader("X-Daytona-Source", "daemon")

	apiClient = apiclient.NewAPIClient(clientConfig)

	apiClient.GetConfig().HTTPClient = &http.Client{
		Transport: http.DefaultTransport,
	}

	return apiClient, nil
}
