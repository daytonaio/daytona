# \WorkspaceConfigAPI

All URIs are relative to *http://localhost:3986*

Method | HTTP request | Description
------------- | ------------- | -------------
[**DeleteWorkspaceConfig**](WorkspaceConfigAPI.md#DeleteWorkspaceConfig) | **Delete** /workspace-config/{configName} | Delete workspace config data
[**GetDefaultWorkspaceConfig**](WorkspaceConfigAPI.md#GetDefaultWorkspaceConfig) | **Get** /workspace-config/default/{gitUrl} | Get workspace configs by git url
[**GetWorkspaceConfig**](WorkspaceConfigAPI.md#GetWorkspaceConfig) | **Get** /workspace-config/{configName} | Get workspace config data
[**ListWorkspaceConfigs**](WorkspaceConfigAPI.md#ListWorkspaceConfigs) | **Get** /workspace-config | List workspace configs
[**SetDefaultWorkspaceConfig**](WorkspaceConfigAPI.md#SetDefaultWorkspaceConfig) | **Patch** /workspace-config/{configName}/set-default | Set workspace config to default
[**SetWorkspaceConfig**](WorkspaceConfigAPI.md#SetWorkspaceConfig) | **Put** /workspace-config | Set workspace config data



## DeleteWorkspaceConfig

> DeleteWorkspaceConfig(ctx, configName).Force(force).Execute()

Delete workspace config data



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/GIT_USER_ID/GIT_REPO_ID/apiclient"
)

func main() {
	configName := "configName_example" // string | Config name
	force := true // bool | Force (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	r, err := apiClient.WorkspaceConfigAPI.DeleteWorkspaceConfig(context.Background(), configName).Force(force).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `WorkspaceConfigAPI.DeleteWorkspaceConfig``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**configName** | **string** | Config name | 

### Other Parameters

Other parameters are passed through a pointer to a apiDeleteWorkspaceConfigRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **force** | **bool** | Force | 

### Return type

 (empty response body)

### Authorization

[Bearer](../README.md#Bearer)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: Not defined

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## GetDefaultWorkspaceConfig

> WorkspaceConfig GetDefaultWorkspaceConfig(ctx, gitUrl).Execute()

Get workspace configs by git url



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/GIT_USER_ID/GIT_REPO_ID/apiclient"
)

func main() {
	gitUrl := "gitUrl_example" // string | Git URL

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.WorkspaceConfigAPI.GetDefaultWorkspaceConfig(context.Background(), gitUrl).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `WorkspaceConfigAPI.GetDefaultWorkspaceConfig``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `GetDefaultWorkspaceConfig`: WorkspaceConfig
	fmt.Fprintf(os.Stdout, "Response from `WorkspaceConfigAPI.GetDefaultWorkspaceConfig`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**gitUrl** | **string** | Git URL | 

### Other Parameters

Other parameters are passed through a pointer to a apiGetDefaultWorkspaceConfigRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


### Return type

[**WorkspaceConfig**](WorkspaceConfig.md)

### Authorization

[Bearer](../README.md#Bearer)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## GetWorkspaceConfig

> WorkspaceConfig GetWorkspaceConfig(ctx, configName).Execute()

Get workspace config data



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/GIT_USER_ID/GIT_REPO_ID/apiclient"
)

func main() {
	configName := "configName_example" // string | Config name

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.WorkspaceConfigAPI.GetWorkspaceConfig(context.Background(), configName).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `WorkspaceConfigAPI.GetWorkspaceConfig``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `GetWorkspaceConfig`: WorkspaceConfig
	fmt.Fprintf(os.Stdout, "Response from `WorkspaceConfigAPI.GetWorkspaceConfig`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**configName** | **string** | Config name | 

### Other Parameters

Other parameters are passed through a pointer to a apiGetWorkspaceConfigRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


### Return type

[**WorkspaceConfig**](WorkspaceConfig.md)

### Authorization

[Bearer](../README.md#Bearer)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: */*

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## ListWorkspaceConfigs

> []WorkspaceConfig ListWorkspaceConfigs(ctx).Execute()

List workspace configs



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/GIT_USER_ID/GIT_REPO_ID/apiclient"
)

func main() {

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.WorkspaceConfigAPI.ListWorkspaceConfigs(context.Background()).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `WorkspaceConfigAPI.ListWorkspaceConfigs``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `ListWorkspaceConfigs`: []WorkspaceConfig
	fmt.Fprintf(os.Stdout, "Response from `WorkspaceConfigAPI.ListWorkspaceConfigs`: %v\n", resp)
}
```

### Path Parameters

This endpoint does not need any parameter.

### Other Parameters

Other parameters are passed through a pointer to a apiListWorkspaceConfigsRequest struct via the builder pattern


### Return type

[**[]WorkspaceConfig**](WorkspaceConfig.md)

### Authorization

[Bearer](../README.md#Bearer)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## SetDefaultWorkspaceConfig

> SetDefaultWorkspaceConfig(ctx, configName).Execute()

Set workspace config to default



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/GIT_USER_ID/GIT_REPO_ID/apiclient"
)

func main() {
	configName := "configName_example" // string | Config name

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	r, err := apiClient.WorkspaceConfigAPI.SetDefaultWorkspaceConfig(context.Background(), configName).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `WorkspaceConfigAPI.SetDefaultWorkspaceConfig``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**configName** | **string** | Config name | 

### Other Parameters

Other parameters are passed through a pointer to a apiSetDefaultWorkspaceConfigRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


### Return type

 (empty response body)

### Authorization

[Bearer](../README.md#Bearer)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: Not defined

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## SetWorkspaceConfig

> SetWorkspaceConfig(ctx).WorkspaceConfig(workspaceConfig).Execute()

Set workspace config data



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/GIT_USER_ID/GIT_REPO_ID/apiclient"
)

func main() {
	workspaceConfig := *openapiclient.NewCreateWorkspaceConfigDTO(map[string]string{"key": "Inner_example"}, "Name_example", "RepositoryUrl_example") // CreateWorkspaceConfigDTO | Workspace config

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	r, err := apiClient.WorkspaceConfigAPI.SetWorkspaceConfig(context.Background()).WorkspaceConfig(workspaceConfig).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `WorkspaceConfigAPI.SetWorkspaceConfig``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiSetWorkspaceConfigRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **workspaceConfig** | [**CreateWorkspaceConfigDTO**](CreateWorkspaceConfigDTO.md) | Workspace config | 

### Return type

 (empty response body)

### Authorization

[Bearer](../README.md#Bearer)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: Not defined

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)

