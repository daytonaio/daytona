# \PluginAPI

All URIs are relative to *http://localhost:3000*

Method | HTTP request | Description
------------- | ------------- | -------------
[**InstallAgentServicePlugin**](PluginAPI.md#InstallAgentServicePlugin) | **Post** /plugin/agent-service/install | Install an agent service plugin
[**InstallProvisionerPlugin**](PluginAPI.md#InstallProvisionerPlugin) | **Post** /plugin/provisioner/install | Install a provisioner plugin
[**ListAgentServicePlugins**](PluginAPI.md#ListAgentServicePlugins) | **Get** /plugin/agent-service | List agent service plugins
[**ListProvisionerPlugins**](PluginAPI.md#ListProvisionerPlugins) | **Get** /plugin/provisioner | List provisioner plugins
[**UninstallAgentServicePlugin**](PluginAPI.md#UninstallAgentServicePlugin) | **Post** /plugin/agent-service/uninstall | Uninstall an agent service plugin
[**UninstallProvisionerPlugin**](PluginAPI.md#UninstallProvisionerPlugin) | **Post** /plugin/provisioner/{provisioner}/uninstall | Uninstall a provisioner plugin



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


## InstallProvisionerPlugin

> InstallProvisionerPlugin(ctx).Plugin(plugin).Execute()

Install a provisioner plugin



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
	r, err := apiClient.PluginAPI.InstallProvisionerPlugin(context.Background()).Plugin(plugin).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `PluginAPI.InstallProvisionerPlugin``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiInstallProvisionerPluginRequest struct via the builder pattern


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


## ListProvisionerPlugins

> []ProvisionerPlugin ListProvisionerPlugins(ctx).Execute()

List provisioner plugins



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
	resp, r, err := apiClient.PluginAPI.ListProvisionerPlugins(context.Background()).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `PluginAPI.ListProvisionerPlugins``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `ListProvisionerPlugins`: []ProvisionerPlugin
	fmt.Fprintf(os.Stdout, "Response from `PluginAPI.ListProvisionerPlugins`: %v\n", resp)
}
```

### Path Parameters

This endpoint does not need any parameter.

### Other Parameters

Other parameters are passed through a pointer to a apiListProvisionerPluginsRequest struct via the builder pattern


### Return type

[**[]ProvisionerPlugin**](ProvisionerPlugin.md)

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


## UninstallProvisionerPlugin

> UninstallProvisionerPlugin(ctx, provisioner).Execute()

Uninstall a provisioner plugin



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
	provisioner := "provisioner_example" // string | Provisioner to uninstall

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	r, err := apiClient.PluginAPI.UninstallProvisionerPlugin(context.Background(), provisioner).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `PluginAPI.UninstallProvisionerPlugin``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**provisioner** | **string** | Provisioner to uninstall | 

### Other Parameters

Other parameters are passed through a pointer to a apiUninstallProvisionerPluginRequest struct via the builder pattern


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

