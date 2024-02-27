# \PluginAPI

All URIs are relative to *http://localhost:3000*

Method | HTTP request | Description
------------- | ------------- | -------------
[**InstallAgentServicePlugin**](PluginAPI.md#InstallAgentServicePlugin) | **Post** /plugin/agent-service/install | Install an agent service plugin
[**InstallProviderPlugin**](PluginAPI.md#InstallProviderPlugin) | **Post** /plugin/provider/install | Install a provider plugin
[**ListAgentServicePlugins**](PluginAPI.md#ListAgentServicePlugins) | **Get** /plugin/agent-service | List agent service plugins
[**ListProviderPlugins**](PluginAPI.md#ListProviderPlugins) | **Get** /plugin/provider | List provider plugins
[**UninstallAgentServicePlugin**](PluginAPI.md#UninstallAgentServicePlugin) | **Post** /plugin/agent-service/uninstall | Uninstall an agent service plugin
[**UninstallProviderPlugin**](PluginAPI.md#UninstallProviderPlugin) | **Post** /plugin/provider/{provider}/uninstall | Uninstall a provider plugin



## InstallAgentServicePlugin

> InstallAgentServicePlugin(ctx).Plugin(plugin).Execute()

Install an agent service plugin



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/GIT_USER_ID/GIT_REPO_ID/serverapiclient"
)

func main() {
	plugin := *openapiclient.NewInstallPluginRequest() // InstallPluginRequest | Plugin to install

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	r, err := apiClient.PluginAPI.InstallAgentServicePlugin(context.Background()).Plugin(plugin).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `PluginAPI.InstallAgentServicePlugin``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiInstallAgentServicePluginRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **plugin** | [**InstallPluginRequest**](InstallPluginRequest.md) | Plugin to install | 

### Return type

 (empty response body)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: Not defined

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## InstallProviderPlugin

> InstallProviderPlugin(ctx).Plugin(plugin).Execute()

Install a provider plugin



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/GIT_USER_ID/GIT_REPO_ID/serverapiclient"
)

func main() {
	plugin := *openapiclient.NewInstallPluginRequest() // InstallPluginRequest | Plugin to install

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	r, err := apiClient.PluginAPI.InstallProviderPlugin(context.Background()).Plugin(plugin).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `PluginAPI.InstallProviderPlugin``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiInstallProviderPluginRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **plugin** | [**InstallPluginRequest**](InstallPluginRequest.md) | Plugin to install | 

### Return type

 (empty response body)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: Not defined

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## ListAgentServicePlugins

> []AgentServicePlugin ListAgentServicePlugins(ctx).Execute()

List agent service plugins



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/GIT_USER_ID/GIT_REPO_ID/serverapiclient"
)

func main() {

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.PluginAPI.ListAgentServicePlugins(context.Background()).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `PluginAPI.ListAgentServicePlugins``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `ListAgentServicePlugins`: []AgentServicePlugin
	fmt.Fprintf(os.Stdout, "Response from `PluginAPI.ListAgentServicePlugins`: %v\n", resp)
}
```

### Path Parameters

This endpoint does not need any parameter.

### Other Parameters

Other parameters are passed through a pointer to a apiListAgentServicePluginsRequest struct via the builder pattern


### Return type

[**[]AgentServicePlugin**](AgentServicePlugin.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## ListProviderPlugins

> []ProviderPlugin ListProviderPlugins(ctx).Execute()

List provider plugins



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/GIT_USER_ID/GIT_REPO_ID/serverapiclient"
)

func main() {

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.PluginAPI.ListProviderPlugins(context.Background()).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `PluginAPI.ListProviderPlugins``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `ListProviderPlugins`: []ProviderPlugin
	fmt.Fprintf(os.Stdout, "Response from `PluginAPI.ListProviderPlugins`: %v\n", resp)
}
```

### Path Parameters

This endpoint does not need any parameter.

### Other Parameters

Other parameters are passed through a pointer to a apiListProviderPluginsRequest struct via the builder pattern


### Return type

[**[]ProviderPlugin**](ProviderPlugin.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## UninstallAgentServicePlugin

> UninstallAgentServicePlugin(ctx, agentService).Execute()

Uninstall an agent service plugin



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/GIT_USER_ID/GIT_REPO_ID/serverapiclient"
)

func main() {
	agentService := "agentService_example" // string | Agent Service to uninstall

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	r, err := apiClient.PluginAPI.UninstallAgentServicePlugin(context.Background(), agentService).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `PluginAPI.UninstallAgentServicePlugin``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**agentService** | **string** | Agent Service to uninstall | 

### Other Parameters

Other parameters are passed through a pointer to a apiUninstallAgentServicePluginRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


### Return type

 (empty response body)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: Not defined

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## UninstallProviderPlugin

> UninstallProviderPlugin(ctx, provider).Execute()

Uninstall a provider plugin



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/GIT_USER_ID/GIT_REPO_ID/serverapiclient"
)

func main() {
	provider := "provider_example" // string | Provider to uninstall

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	r, err := apiClient.PluginAPI.UninstallProviderPlugin(context.Background(), provider).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `PluginAPI.UninstallProviderPlugin``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**provider** | **string** | Provider to uninstall | 

### Other Parameters

Other parameters are passed through a pointer to a apiUninstallProviderPluginRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


### Return type

 (empty response body)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: Not defined

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)

