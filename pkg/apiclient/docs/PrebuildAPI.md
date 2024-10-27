# \PrebuildAPI

All URIs are relative to *http://localhost:3986*

Method | HTTP request | Description
------------- | ------------- | -------------
[**DeletePrebuild**](PrebuildAPI.md#DeletePrebuild) | **Delete** /workspace-config/{configName}/prebuild/{prebuildId} | Delete prebuild
[**GetPrebuild**](PrebuildAPI.md#GetPrebuild) | **Get** /workspace-config/{configName}/prebuild/{prebuildId} | Get prebuild
[**ListPrebuilds**](PrebuildAPI.md#ListPrebuilds) | **Get** /workspace-config/prebuild | List prebuilds
[**ListPrebuildsForWorkspaceConfig**](PrebuildAPI.md#ListPrebuildsForWorkspaceConfig) | **Get** /workspace-config/{configName}/prebuild | List prebuilds for workspace config
[**ProcessGitEvent**](PrebuildAPI.md#ProcessGitEvent) | **Post** /workspace-config/prebuild/process-git-event | ProcessGitEvent
[**SetPrebuild**](PrebuildAPI.md#SetPrebuild) | **Put** /workspace-config/{configName}/prebuild | Set prebuild



## DeletePrebuild

> DeletePrebuild(ctx, configName, prebuildId).Force(force).Execute()

Delete prebuild



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
	configName := "configName_example" // string | Workspace config name
	prebuildId := "prebuildId_example" // string | Prebuild ID
	force := true // bool | Force (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	r, err := apiClient.PrebuildAPI.DeletePrebuild(context.Background(), configName, prebuildId).Force(force).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `PrebuildAPI.DeletePrebuild``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**configName** | **string** | Workspace config name | 
**prebuildId** | **string** | Prebuild ID | 

### Other Parameters

Other parameters are passed through a pointer to a apiDeletePrebuildRequest struct via the builder pattern


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


## GetPrebuild

> PrebuildDTO GetPrebuild(ctx, configName, prebuildId).Execute()

Get prebuild



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
	configName := "configName_example" // string | Workspace config name
	prebuildId := "prebuildId_example" // string | Prebuild ID

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.PrebuildAPI.GetPrebuild(context.Background(), configName, prebuildId).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `PrebuildAPI.GetPrebuild``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `GetPrebuild`: PrebuildDTO
	fmt.Fprintf(os.Stdout, "Response from `PrebuildAPI.GetPrebuild`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**configName** | **string** | Workspace config name | 
**prebuildId** | **string** | Prebuild ID | 

### Other Parameters

Other parameters are passed through a pointer to a apiGetPrebuildRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------



### Return type

[**PrebuildDTO**](PrebuildDTO.md)

### Authorization

[Bearer](../README.md#Bearer)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: */*

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## ListPrebuilds

> []PrebuildDTO ListPrebuilds(ctx).Execute()

List prebuilds



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
	resp, r, err := apiClient.PrebuildAPI.ListPrebuilds(context.Background()).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `PrebuildAPI.ListPrebuilds``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `ListPrebuilds`: []PrebuildDTO
	fmt.Fprintf(os.Stdout, "Response from `PrebuildAPI.ListPrebuilds`: %v\n", resp)
}
```

### Path Parameters

This endpoint does not need any parameter.

### Other Parameters

Other parameters are passed through a pointer to a apiListPrebuildsRequest struct via the builder pattern


### Return type

[**[]PrebuildDTO**](PrebuildDTO.md)

### Authorization

[Bearer](../README.md#Bearer)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: */*

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## ListPrebuildsForWorkspaceConfig

> []PrebuildDTO ListPrebuildsForWorkspaceConfig(ctx, configName).Execute()

List prebuilds for workspace config



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
	resp, r, err := apiClient.PrebuildAPI.ListPrebuildsForWorkspaceConfig(context.Background(), configName).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `PrebuildAPI.ListPrebuildsForWorkspaceConfig``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `ListPrebuildsForWorkspaceConfig`: []PrebuildDTO
	fmt.Fprintf(os.Stdout, "Response from `PrebuildAPI.ListPrebuildsForWorkspaceConfig`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**configName** | **string** | Config name | 

### Other Parameters

Other parameters are passed through a pointer to a apiListPrebuildsForWorkspaceConfigRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


### Return type

[**[]PrebuildDTO**](PrebuildDTO.md)

### Authorization

[Bearer](../README.md#Bearer)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: */*

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## ProcessGitEvent

> ProcessGitEvent(ctx).Body(body).Execute()

ProcessGitEvent



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
	body := map[string]interface{}{ ... } // map[string]interface{} | Webhook event

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	r, err := apiClient.PrebuildAPI.ProcessGitEvent(context.Background()).Body(body).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `PrebuildAPI.ProcessGitEvent``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiProcessGitEventRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **body** | **map[string]interface{}** | Webhook event | 

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


## SetPrebuild

> string SetPrebuild(ctx, configName).Prebuild(prebuild).Execute()

Set prebuild



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
	prebuild := *openapiclient.NewCreatePrebuildDTO(int32(123)) // CreatePrebuildDTO | Prebuild

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.PrebuildAPI.SetPrebuild(context.Background(), configName).Prebuild(prebuild).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `PrebuildAPI.SetPrebuild``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `SetPrebuild`: string
	fmt.Fprintf(os.Stdout, "Response from `PrebuildAPI.SetPrebuild`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**configName** | **string** | Config name | 

### Other Parameters

Other parameters are passed through a pointer to a apiSetPrebuildRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **prebuild** | [**CreatePrebuildDTO**](CreatePrebuildDTO.md) | Prebuild | 

### Return type

**string**

### Authorization

[Bearer](../README.md#Bearer)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: */*

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)

