# \PrebuildAPI

All URIs are relative to *http://localhost:3986*

Method | HTTP request | Description
------------- | ------------- | -------------
[**DeletePrebuild**](PrebuildAPI.md#DeletePrebuild) | **Delete** /project-config/prebuild | Delete prebuild
[**ListPrebuilds**](PrebuildAPI.md#ListPrebuilds) | **Get** /project-config/prebuild | List prebuilds
[**ProcessGitEvent**](PrebuildAPI.md#ProcessGitEvent) | **Post** /project-config/prebuild/process-git-event | ProcessGitEvent
[**SetPrebuild**](PrebuildAPI.md#SetPrebuild) | **Post** /project-config/prebuild | Set prebuild



## DeletePrebuild

> DeletePrebuild(ctx).Prebuild(prebuild).Execute()

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
	prebuild := *openapiclient.NewPrebuildConfig() // PrebuildConfig | Prebuild

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	r, err := apiClient.PrebuildAPI.DeletePrebuild(context.Background()).Prebuild(prebuild).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `PrebuildAPI.DeletePrebuild``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiDeletePrebuildRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **prebuild** | [**PrebuildConfig**](PrebuildConfig.md) | Prebuild | 

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


## ListPrebuilds

> []PrebuildDTO ListPrebuilds(ctx, configName).Execute()

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
	configName := "configName_example" // string | Config name

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.PrebuildAPI.ListPrebuilds(context.Background(), configName).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `PrebuildAPI.ListPrebuilds``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `ListPrebuilds`: []PrebuildDTO
	fmt.Fprintf(os.Stdout, "Response from `PrebuildAPI.ListPrebuilds`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**configName** | **string** | Config name | 

### Other Parameters

Other parameters are passed through a pointer to a apiListPrebuildsRequest struct via the builder pattern


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

> ProcessGitEvent(ctx).Workspace(workspace).Execute()

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
	workspace := map[string]interface{}{ ... } // map[string]interface{} | Webhook event

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	r, err := apiClient.PrebuildAPI.ProcessGitEvent(context.Background()).Workspace(workspace).Execute()
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
 **workspace** | **map[string]interface{}** | Webhook event | 

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

> SetPrebuild(ctx).Prebuild(prebuild).Execute()

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
	prebuild := *openapiclient.NewCreatePrebuildDTO() // CreatePrebuildDTO | Prebuild

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	r, err := apiClient.PrebuildAPI.SetPrebuild(context.Background()).Prebuild(prebuild).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `PrebuildAPI.SetPrebuild``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiSetPrebuildRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **prebuild** | [**CreatePrebuildDTO**](CreatePrebuildDTO.md) | Prebuild | 

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

